// Package med66 implements source-aligned Med66 course extraction.
package med66

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/nichuanfang/medigo/internal/extractor"
	"github.com/nichuanfang/medigo/internal/extractor/shared"
	"github.com/nichuanfang/medigo/internal/util"
)

const (
	LOGIN_URL            = "https://www.med66.com/OtherItem/loginAgain/index.shtml"
	MEMBER_HOME_URL      = "https://member.med66.com/homes/mycourse"
	COURSE_INFO_URL      = "https://member.med66.com/homes/mycourse/courseInfo"
	COURSEWARE_INFO_URL  = "https://member.med66.com/homes/course/courseClassWareInfo"
	ELEARNING_HOME_URL   = "https://elearning.med66.com/"
	LIVE_REPLAY_INFO_URL = "https://live.cdeledu.com/liveapi/entry/getReplayInfo"
	LIVE_REFERER_URL     = "https://live.cdeledu.com/"
)

var patterns = []string{`(?:[\w-]+\.)?med66\.com/`, `live\.cdeledu\.com/`}

func init() {
	extractor.Register(&Med66{}, extractor.SiteInfo{Name: "Med66", URL: "med66.com", NeedAuth: true})
}

type Med66 struct{}

func (m *Med66) Patterns() []string { return patterns }

var (
	courseIDRe = regexp.MustCompile(`(?i)(?:courseId|course_id)=((?:med)?\d+)`)
	goToLiveRe = regexp.MustCompile(`goToLive\(\s*['"]([^'"]+)['"](?:\s*,\s*['"]([^'"]*)['"])?`)
)

func (m *Med66) Extract(rawURL string, opts *extractor.ExtractOpts) (*extractor.MediaInfo, error) {
	if opts == nil || opts.Cookies == nil {
		return nil, fmt.Errorf("med66 requires login cookies")
	}

	c := util.NewClient()
	c.SetCookieJar(opts.Cookies)
	headers := med66Headers()
	uid := cookieValue(opts.Cookies, []string{"https://member.med66.com/", "https://www.med66.com/"}, "cdeluid")

	if isReplayURL(rawURL) {
		entry, err := resolveReplayEntry(c, rawURL, rawURL, uid, "med66_replay")
		if err != nil {
			return nil, err
		}
		return entry, nil
	}

	cid := extractCourseID(rawURL)
	if cid == "" {
		return nil, fmt.Errorf("cannot parse med66 courseId from URL: %s", rawURL)
	}

	course, err := fetchCourse(c, headers, cid)
	if err != nil {
		return nil, err
	}
	wares, err := fetchCoursewares(c, headers, course)
	if err != nil {
		return nil, err
	}

	var entries []*extractor.MediaInfo
	for i, ware := range wares {
		if !isRecordedWare(ware) {
			continue
		}
		pageURL := normalizeURL(firstString(ware, "cwDirURL", "dirURL", "cwURL"), ELEARNING_HOME_URL)
		if pageURL == "" {
			continue
		}
		body, err := c.GetString(pageURL, map[string]string{"Referer": MEMBER_HOME_URL})
		if err != nil {
			continue
		}
		matches := goToLiveRe.FindAllStringSubmatch(body, -1)
		for j, match := range matches {
			playURL := normalizeURL(match[1], ELEARNING_HOME_URL)
			if playURL == "" {
				continue
			}
			title := fmt.Sprintf("%02d.%02d %s", i+1, j+1, titleFromWare(ware))
			entry, err := resolveReplayEntry(c, playURL, pageURL, uid, title)
			if err == nil && entry != nil {
				entry.Extra["course_id"] = course.CourseID
				entry.Extra["cware_id"] = firstString(ware, "cwareId", "cwareID", "cware_id")
				entries = append(entries, entry)
			}
		}
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("med66: no playable live replay entries found (course locked or schema changed)")
	}

	return &extractor.MediaInfo{Site: "med66", Title: course.Title, Entries: entries}, nil
}

type med66Course struct {
	CourseID        string
	Title           string
	EduSubjectID    string
	ClassType       string
	ClassID         string
	LinkedCourseIDs string
}

type anyMap map[string]any

func fetchCourse(c *util.Client, headers map[string]string, cid string) (med66Course, error) {
	body, err := c.PostForm(COURSE_INFO_URL, map[string]string{}, headers)
	if err != nil {
		return med66Course{}, fmt.Errorf("courseInfo: %w", err)
	}
	var root any
	if err := json.Unmarshal([]byte(body), &root); err != nil {
		return med66Course{}, fmt.Errorf("parse courseInfo: %w", err)
	}
	for _, obj := range collectMaps(root) {
		courseID := strAny(obj["courseId"])
		if courseID == "" {
			courseID = strAny(obj["course_id"])
		}
		if courseID != cid {
			continue
		}
		return med66Course{
			CourseID:        courseID,
			Title:           firstString(obj, "title", "homeTitle", "selCourseTitle", "courseEduName", "listName"),
			EduSubjectID:    firstString(obj, "eduSubjectId", "eduSubjectID"),
			ClassType:       firstString(obj, "classType"),
			ClassID:         firstString(obj, "classId", "viewClassId"),
			LinkedCourseIDs: firstString(obj, "linkedCourseIds"),
		}, nil
	}
	return med66Course{}, fmt.Errorf("med66 courseInfo: courseId %s not found", cid)
}

func fetchCoursewares(c *util.Client, headers map[string]string, course med66Course) ([]anyMap, error) {
	form := map[string]string{
		"eduSubjectId":    course.EduSubjectID,
		"classType":       course.ClassType,
		"classId":         course.ClassID,
		"linkedCourseIds": course.LinkedCourseIDs,
		"courseId":        course.CourseID,
	}
	body, err := c.PostForm(COURSEWARE_INFO_URL, form, headers)
	if err != nil {
		return nil, fmt.Errorf("courseClassWareInfo: %w", err)
	}
	var root any
	if err := json.Unmarshal([]byte(body), &root); err != nil {
		return nil, fmt.Errorf("parse courseClassWareInfo: %w", err)
	}
	var out []anyMap
	for _, obj := range collectMaps(root) {
		for _, key := range []string{"homeCwareList", "homeWareList", "courseWareList", "wareList"} {
			if list, ok := obj[key].([]any); ok {
				for _, item := range list {
					if m, ok := item.(map[string]any); ok {
						out = append(out, anyMap(m))
					}
				}
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("med66: empty wareList for courseId %s", course.CourseID)
	}
	return out, nil
}

func resolveReplayEntry(c *util.Client, playURL, referer, uid, title string) (*extractor.MediaInfo, error) {
	payload, err := resolveReplayPayload(c, playURL, referer)
	if err != nil {
		return nil, err
	}
	replay := payload.Replay
	viewer := strings.TrimSpace(uid)
	if viewer == "" {
		viewer = payload.Query.Get("userid")
	}
	candidates := uniqueNonEmpty(replay.AccessKey, payload.Token)
	if len(candidates) == 0 {
		candidates = []string{""}
	}

	var playInfo *shared.CssLcloudPlayInfo
	var lastErr error
	for _, token := range candidates {
		playInfo, lastErr = shared.CssLcloudResolvePlayInfo(c, shared.CssLcloudPayload{
			LiveRoomID:  firstNonEmpty(replay.LiveRoomID, replay.LiveID),
			UserID:      replay.AccessID,
			AccessID:    replay.AccessID,
			RecordID:    replay.RecordID,
			ViewerName:  viewer,
			ViewerToken: token,
			Referer:     LIVE_REFERER_URL,
		})
		if lastErr == nil {
			break
		}
	}
	if playInfo == nil || lastErr != nil {
		return nil, fmt.Errorf("med66 csslcloud replay: %w", lastErr)
	}

	streams := map[string]extractor.Stream{}
	for idx, s := range playInfo.VideoList {
		if s.URL == "" {
			continue
		}
		key := fmt.Sprintf("definition_%d", s.Definition)
		if s.Definition == 0 {
			key = fmt.Sprintf("stream_%d", idx+1)
		}
		streams[key] = extractor.Stream{Quality: key, URLs: []string{s.URL}, Format: pickFormat(s.URL), AudioURL: playInfo.AudioURL, Headers: map[string]string{"Referer": LIVE_REFERER_URL}}
	}
	if len(streams) == 0 && playInfo.VideoURL != "" {
		streams["best"] = extractor.Stream{Quality: "best", URLs: []string{playInfo.VideoURL}, Format: pickFormat(playInfo.VideoURL), AudioURL: playInfo.AudioURL, Headers: map[string]string{"Referer": LIVE_REFERER_URL}}
	}
	if len(streams) == 0 {
		return nil, fmt.Errorf("med66 csslcloud: no media URL")
	}

	extra := map[string]any{"recordId": replay.RecordID, "accessid": replay.AccessID, "liveRoomId": firstNonEmpty(replay.LiveRoomID, replay.LiveID)}
	if strings.Contains(playInfo.VideoURL, ".m3u8") {
		if text, err := c.GetString(playInfo.VideoURL, map[string]string{"Referer": LIVE_REFERER_URL}); err == nil {
			if rewritten, err := shared.CssLcloudRewriteM3U8Keys(c, text, LIVE_REFERER_URL); err == nil {
				extra["m3u8_text"] = rewritten
			}
		}
	}
	return &extractor.MediaInfo{Site: "med66", Title: title, Streams: streams, Extra: extra}, nil
}

type liveReplayPayload struct {
	Replay liveReplayReplay
	Token  string
	Query  url.Values
}

type liveReplayReplay struct {
	LiveRoomID string `json:"liveRoomId"`
	LiveID     string `json:"liveId"`
	AccessID   string `json:"accessid"`
	RecordID   string `json:"recordId"`
	AccessKey  string `json:"accesskey"`
}

func resolveReplayPayload(c *util.Client, playURL, referer string) (liveReplayPayload, error) {
	finalURL := playURL
	if !strings.Contains(playURL, "liveapi/entry/getReplayInfo") {
		resp, err := c.Get(playURL, map[string]string{"Referer": referer})
		if err != nil {
			return liveReplayPayload{}, fmt.Errorf("resolve replay redirect: %w", err)
		}
		resp.Body.Close()
		finalURL = resp.Request.URL.String()
	}
	query, err := stripReplayQuery(finalURL)
	if err != nil {
		return liveReplayPayload{}, err
	}
	body, err := getJSONWithParams(c, LIVE_REPLAY_INFO_URL, query, map[string]string{"Referer": LIVE_REFERER_URL})
	if err != nil {
		return liveReplayPayload{}, fmt.Errorf("getReplayInfo: %w", err)
	}
	var root struct {
		Replay liveReplayReplay `json:"replay"`
		Token  string           `json:"token"`
		Data   struct {
			Replay liveReplayReplay `json:"replay"`
			Token  string           `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &root); err != nil {
		return liveReplayPayload{}, fmt.Errorf("parse getReplayInfo: %w", err)
	}
	replay := root.Replay
	if replay.RecordID == "" {
		replay = root.Data.Replay
	}
	token := firstNonEmpty(root.Token, root.Data.Token)
	if firstNonEmpty(replay.LiveRoomID, replay.LiveID) == "" || replay.AccessID == "" || replay.RecordID == "" {
		return liveReplayPayload{}, fmt.Errorf("med66 getReplayInfo: missing replay liveRoomId/accessid/recordId")
	}
	return liveReplayPayload{Replay: replay, Token: token, Query: query}, nil
}

func getJSONWithParams(c *util.Client, endpoint string, params url.Values, headers map[string]string) (string, error) {
	u, _ := url.Parse(endpoint)
	u.RawQuery = params.Encode()
	return c.GetString(u.String(), headers)
}

func stripReplayQuery(s string) (url.Values, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Del("oldTime")
	q.Del("oldKey")
	return q, nil
}

func med66Headers() map[string]string {
	return map[string]string{"Origin": "https://member.med66.com", "Referer": MEMBER_HOME_URL, "Accept": "application/json, text/plain, */*"}
}

func isReplayURL(s string) bool {
	return strings.Contains(s, "live.cdeledu.com") || strings.Contains(s, "recordId=") || strings.Contains(s, "liveRoomId=")
}

func extractCourseID(s string) string {
	if m := courseIDRe.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	return ""
}

func isRecordedWare(m anyMap) bool {
	return firstString(m, "cwDirURL", "dirURL", "cwURL", "cwareId", "cwareID") != ""
}

func titleFromWare(m anyMap) string {
	if t := firstString(m, "cwName", "cwShowName", "title"); t != "" {
		return util.SanitizeFilename(t)
	}
	return "课件"
}
