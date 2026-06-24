// Package houda implements an extractor for houdask.com courses that play via CSSLCloud.
//
// API endpoints from decompiled Mooc/Courses/Houda/Houda_Course.pyc:
//
//	http://www.houdask.com/api/center/online/myOnline/anon/getLearnFirstPage
//	http://www.houdask.com/api/center/online/myOnline/getXxStageAndLawList
//	http://www.houdask.com/api/center/myOnlineCourse/getLearnCourse
//	http://www.houdask.com/api/center/myOnlineCourse/getLearnCoursePage
//	http://www.houdask.com/api/center/myOnlineLive/anon/v202401/getById
//	http://www.houdask.com/api/center/myLibraryMaterial/v2/getList
//	http://www.houdask.com/api/center/live/cc/anon/viewPlayback/{room_id}/{record_id}
//	https://view.csslcloud.net/replay/user/login
//	https://view.csslcloud.net/replay/video/play
//	https://view.csslcloud.net/replay/data/meta
package houda

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/nichuanfang/medigo/internal/extractor"
	"github.com/nichuanfang/medigo/internal/extractor/shared"
	"github.com/nichuanfang/medigo/internal/util"
)

const (
	urlOrigin          = "http://www.houdask.com"
	urlHome            = "http://www.houdask.com/"
	urlLoginCheck      = "http://www.houdask.com/api/center/sysUserPower/anon/ifLogin"
	urlCourseList      = "http://www.houdask.com/api/center/online/myOnline/anon/getLearnFirstPage"
	urlStageLaw        = "http://www.houdask.com/api/center/online/myOnline/getXxStageAndLawList"
	urlLearnCourse     = "http://www.houdask.com/api/center/myOnlineCourse/getLearnCourse"
	urlLearnCoursePage = "http://www.houdask.com/api/center/myOnlineCourse/getLearnCoursePage"
	urlLiveDetail      = "http://www.houdask.com/api/center/myOnlineLive/anon/v202401/getById"
	urlMaterial        = "http://www.houdask.com/api/center/myLibraryMaterial/v2/getList"
	urlCCViewPlayback  = "http://www.houdask.com/api/center/live/cc/anon/viewPlayback/%s/%s"
	urlCsslLogin       = "https://view.csslcloud.net/replay/user/login"
	urlCsslPlay        = "https://view.csslcloud.net/replay/video/play"
	urlCsslMeta        = "https://view.csslcloud.net/replay/data/meta"
	urlCsslOrigin      = "https://view.csslcloud.net"
	csslDeviceType     = "h5-pc"
	csslDeviceVersion  = "3.21.0"
	csslTpl            = "20"
	csslTerminal       = "3"
	materialServiceTyp = "1"
)

var patterns = []string{
	`(?:[\w-]+\.)?houdask\.com/`,
	`(?:[\w-]+\.)?csslcloud\.net/`,
}

func init() {
	extractor.Register(&Houda{}, extractor.SiteInfo{Name: "Houda", URL: "houdask.com", NeedAuth: true})
}

type Houda struct{}

func (s *Houda) Patterns() []string { return patterns }

var classIDRe = regexp.MustCompile(`(?i)(?:classId|class_id|courseId|course_id|id)=([0-9]+)|/(?:online|course|class|learn|myOnline)[^?#/]*/([0-9]+)(?:[/?#]|$)`)

func (s *Houda) Extract(rawURL string, opts *extractor.ExtractOpts) (*extractor.MediaInfo, error) {
	if opts == nil || opts.Cookies == nil {
		return nil, fmt.Errorf("houda requires login cookies")
	}
	c := util.NewClient()
	c.SetCookieJar(opts.Cookies)
	headers := houdaHeaders()

	if isCsslcloudURL(rawURL) {
		return s.extractCsslcloudURL(c, rawURL, headers)
	}
	if err := checkHoudaCookie(c, headers); err != nil {
		return nil, err
	}
	cid := parseClassID(rawURL)
	courses, _ := fetchHoudaCourseList(c, headers)
	course := chooseHoudaCourse(courses, cid, rawURL)
	if cid == "" {
		cid = course.ID
	}
	if cid == "" {
		return nil, fmt.Errorf("cannot parse houda classId from URL and course list is empty: %s", rawURL)
	}
	lessons, err := fetchHoudaLessons(c, cid, headers)
	if err != nil {
		return nil, err
	}
	stageLaw, _ := fetchHoudaStageLaw(c, cid, headers)

	entries := make([]*extractor.MediaInfo, 0, len(lessons))
	for i, lesson := range lessons {
		entry, err := buildHoudaEntry(c, cid, i+1, lesson, headers)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	materials := fetchHoudaMaterials(c, cid, stageLaw, headers)
	for i, material := range materials {
		if entry := buildHoudaMaterialEntry(i+1, material, headers); entry != nil {
			entries = append(entries, entry)
		}
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("houda: no playable video or material entries for classId=%s", cid)
	}
	title := firstNonEmpty(course.Title, "houda_"+cid)
	return &extractor.MediaInfo{Site: "houda", Title: title, Entries: entries, Extra: map[string]any{"course_id": cid, "course": course.Raw}}, nil
}

func houdaHeaders() map[string]string {
	return map[string]string{
		"appType":          "WEB",
		"X-Requested-With": "XMLHttpRequest",
		"Accept":           "application/json, text/plain, */*",
		"Origin":           urlOrigin,
		"Referer":          urlHome,
	}
}

func checkHoudaCookie(c *util.Client, headers map[string]string) error {
	body, err := c.GetString(urlLoginCheck, headers)
	if err != nil {
		return fmt.Errorf("houda cookie check: %w", err)
	}
	var out struct {
		Code any `json:"code"`
		Data any `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		return fmt.Errorf("houda cookie check parse: %w", err)
	}
	if stringValue(out.Code) != "1" {
		return fmt.Errorf("houda requires valid logged-in cookie (code=%s)", stringValue(out.Code))
	}
	return nil
}

func fetchHoudaLessons(c *util.Client, cid string, headers map[string]string) ([]houdaLesson, error) {
	payloads := []struct {
		endpoint string
		data     map[string]string
	}{
		{urlLearnCourse, map[string]string{"classId": cid}},
		{urlLearnCoursePage, map[string]string{"classId": cid, "pageNum": "1", "pageSize": "500", "page": "1", "size": "500"}},
	}
	var lastErr error
	for _, req := range payloads {
		root, err := requestHouda(c, req.endpoint, req.data, headers)
		if err != nil {
			lastErr = err
			continue
		}
		lessons := parseHoudaLessons(root)
		if len(lessons) > 0 {
			return lessons, nil
		}
		lastErr = fmt.Errorf("empty liveList from %s", req.endpoint)
	}
	return nil, fmt.Errorf("houda fetch lessons: %w", lastErr)
}

func requestHouda(c *util.Client, endpoint string, data map[string]string, headers map[string]string) (map[string]any, error) {
	if data == nil {
		data = map[string]string{}
	}
	wrappedBytes, _ := json.Marshal(data)
	forms := []map[string]string{
		data,
		{"data": string(wrappedBytes)},
	}
	var lastErr error
	for _, form := range forms {
		body, err := c.PostForm(endpoint, form, headers)
		if err != nil {
			lastErr = err
			continue
		}
		var root map[string]any
		if err := json.Unmarshal([]byte(body), &root); err != nil {
			lastErr = err
			continue
		}
		return root, nil
	}
	return nil, lastErr
}

type houdaCourse struct {
	ID    string
	Title string
	Raw   map[string]any
}

type houdaStageLaw struct {
	Raw  map[string]any
	Laws []houdaLawRef
}

type houdaLawRef struct {
	ID    string
	Title string
}

func fetchHoudaCourseList(c *util.Client, headers map[string]string) ([]houdaCourse, error) {
	root, err := requestHouda(c, urlCourseList, map[string]string{}, headers)
	if err != nil {
		return nil, err
	}
	return parseHoudaCourseList(root), nil
}

func parseHoudaCourseList(root map[string]any) []houdaCourse {
	data := unwrapHoudaData(root)
	var courses []houdaCourse
	seen := map[string]bool{}
	tabs := houdaMapList(data, "tabList", "tabs", "list", "items")
	if len(tabs) == 0 {
		tabs = []map[string]any{{"dataList": data}}
	}
	for _, tab := range tabs {
		tabName := firstMapText(tab, "name", "tabName", "title")
		code := strings.ToUpper(firstMapText(tab, "code"))
		if strings.Contains(tabName, "资料") || code == "ZL" {
			continue
		}
		rows := houdaMapList(tab, "dataList", "list", "items", "courseList", "courses")
		if len(rows) == 0 && hasAnyMap(tab, "id", "classId", "courseId") {
			rows = []map[string]any{tab}
		}
		for _, row := range rows {
			id := firstMapText(row, "id", "classId", "courseId", "class_id", "course_id")
			title := firstMapText(row, "name", "title", "courseName", "course_name")
			if id == "" || title == "" || seen[id] {
				continue
			}
			seen[id] = true
			raw := cloneMap(row)
			raw["tab_name"] = tabName
			courses = append(courses, houdaCourse{ID: id, Title: title, Raw: raw})
		}
	}
	return courses
}

func chooseHoudaCourse(courses []houdaCourse, cid, rawURL string) houdaCourse {
	if cid != "" {
		for _, course := range courses {
			if course.ID == cid {
				return course
			}
		}
		return houdaCourse{ID: cid, Title: "houda_" + cid}
	}
	if u, err := url.Parse(rawURL); err == nil {
		nameHint, _ := url.QueryUnescape(u.Query().Get("name"))
		nameHint = strings.TrimSpace(nameHint)
		if nameHint != "" {
			for _, course := range courses {
				if strings.Contains(course.Title, nameHint) || strings.Contains(nameHint, course.Title) {
					return course
				}
			}
		}
	}
	if len(courses) > 0 {
		return courses[0]
	}
	return houdaCourse{}
}

func fetchHoudaStageLaw(c *util.Client, cid string, headers map[string]string) (*houdaStageLaw, error) {
	root, err := requestHouda(c, urlStageLaw, map[string]string{"classId": cid}, headers)
	if err != nil {
		return nil, err
	}
	raw := unwrapHoudaData(root)
	stageLaw := &houdaStageLaw{Raw: raw, Laws: collectHoudaLawRefs(raw)}
	return stageLaw, nil
}

func fetchHoudaMaterials(c *util.Client, cid string, stageLaw *houdaStageLaw, headers map[string]string) []map[string]any {
	lawIDs := []string{""}
	if stageLaw != nil {
		for _, law := range stageLaw.Laws {
			lawIDs = appendStringUnique(lawIDs, law.ID)
		}
	}
	var out []map[string]any
	seen := map[string]bool{}
	for _, lawID := range lawIDs {
		root, err := requestHouda(c, urlMaterial, map[string]string{"lawId": lawID, "serviceType": materialServiceTyp, "classId": cid}, headers)
		if err != nil {
			continue
		}
		rows := houdaMapList(unwrapHoudaData(root), "data", "list", "items", "rows", "records")
		for _, row := range rows {
			key := houdaMaterialKey(row)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, row)
		}
	}
	return out
}

func hydrateHoudaLesson(c *util.Client, lesson houdaLesson, headers map[string]string) houdaLesson {
	lessonID := firstText(lesson.ID)
	if lessonID == "" {
		return lesson
	}
	roomID := firstText(lesson.RoomID, lesson.MainRoomID, lesson.CCLiveID)
	recordID := firstText(lesson.RecordID)
	direct := firstText(lesson.PlaybackMP4, lesson.PlaybackURL, lesson.LiveURL)
	if roomID != "" && (recordID != "" || direct != "") {
		return lesson
	}
	detail, err := fetchHoudaLiveDetail(c, lessonID, headers)
	if err != nil {
		return lesson
	}
	return mergeHoudaLesson(lesson, detail)
}

func fetchHoudaLiveDetail(c *util.Client, lessonID string, headers map[string]string) (houdaLesson, error) {
	for _, data := range []map[string]string{{"id": lessonID}, {"liveId": lessonID}} {
		root, err := requestHouda(c, urlLiveDetail, data, headers)
		if err != nil {
			continue
		}
		if lesson := houdaLessonFromAny(unwrapHoudaData(root)); firstText(lesson.ID, lesson.RecordID, lesson.RoomID, lesson.PlaybackURL, lesson.PlaybackMP4) != "" {
			if firstText(lesson.ID) == "" {
				lesson.ID = lessonID
			}
			return lesson, nil
		}
	}
	api := addHoudaQuery(urlLiveDetail, map[string]string{"id": lessonID})
	body, err := c.GetString(api, headers)
	if err != nil {
		return houdaLesson{}, err
	}
	var root map[string]any
	if err := json.Unmarshal([]byte(body), &root); err != nil {
		return houdaLesson{}, err
	}
	lesson := houdaLessonFromAny(unwrapHoudaData(root))
	if firstText(lesson.ID) == "" {
		lesson.ID = lessonID
	}
	return lesson, nil
}

func buildHoudaMaterialEntry(index int, item map[string]any, headers map[string]string) *extractor.MediaInfo {
	fileURL := normalizeHoudaURL(firstMapText(item, "downLoadUrl", "downloadUrl", "fileUrl", "url", "path"))
	if fileURL == "" {
		return nil
	}
	name := firstMapText(item, "title", "name", "fileName", "materialName", "coursewareName")
	if name == "" {
		name = "资料"
	}
	format := houdaFileExt(firstMapText(item, "fileType", "type", "ext", "format"), fileURL)
	return &extractor.MediaInfo{
		Site:  "houda",
		Title: fmt.Sprintf("(%d)--%s", index, name),
		Streams: map[string]extractor.Stream{"file": {
			Quality: "file",
			URLs:    []string{fileURL},
			Format:  format,
			Headers: map[string]string{"Referer": urlHome, "appType": "WEB"},
		}},
		Extra: map[string]any{"kind": "material", "raw": item},
	}
}

type houdaLesson struct {
	ID          any `json:"id"`
	Title       any `json:"title"`
	Name        any `json:"name"`
	CourseName  any `json:"courseName"`
	CourseID    any `json:"courseId"`
	ClassID     any `json:"classId"`
	Type        any `json:"type"`
	CCLiveID    any `json:"ccLiveId"`
	RoomID      any `json:"roomId"`
	MainRoomID  any `json:"mainRoomId"`
	RecordID    any `json:"recordId"`
	LiveURL     any `json:"liveUrl"`
	PlaybackURL any `json:"playbackUrl"`
	PlaybackMP4 any `json:"playbackMp4"`
	PlaybackMP3 any `json:"playbackMp3"`
	StageID     any `json:"stageId"`
	StageName   any `json:"stageName"`
	LawID       any `json:"lawId"`
	LawName     any `json:"lawName"`
}

func buildHoudaEntry(c *util.Client, cid string, index int, lesson houdaLesson, headers map[string]string) (*extractor.MediaInfo, error) {
	lesson = hydrateHoudaLesson(c, lesson, headers)
	title := firstText(lesson.Title, lesson.Name, lesson.CourseName, "未命名")
	lessonID := firstText(lesson.ID)
	roomID := firstText(lesson.RoomID, lesson.MainRoomID, lesson.CCLiveID)
	recordID := firstText(lesson.RecordID)
	direct := firstText(lesson.PlaybackMP4, lesson.PlaybackURL, lesson.LiveURL)

	streams := map[string]extractor.Stream{}
	extra := map[string]any{"course_id": cid, "lesson_id": lessonID, "record_id": recordID, "room_id": roomID}
	if roomID != "" && recordID != "" {
		info, err := resolveHoudaCSSL(c, roomID, recordID, title, headers)
		if err == nil {
			extra["csslcloud_session_id"] = info.SessionID
			for _, v := range info.VideoList {
				if v.URL == "" {
					continue
				}
				key := fmt.Sprintf("definition_%d", v.Definition)
				streams[key] = extractor.Stream{Quality: key, URLs: []string{v.URL}, Format: mediaExt(v.URL), AudioURL: info.AudioURL, Headers: map[string]string{"Referer": urlHome}}
			}
			if info.VideoURL != "" && mediaExt(info.VideoURL) == "m3u8" {
				if rewritten, err := rewriteHoudaM3U8(c, info.VideoURL, urlHome); err == nil {
					extra["m3u8_text"] = rewritten
				}
			}
		} else if direct == "" {
			return nil, err
		} else {
			extra["csslcloud_error"] = err.Error()
		}
	}
	if len(streams) == 0 && direct != "" {
		fmtName := mediaExt(direct)
		streams[fmtName] = extractor.Stream{Quality: "best", URLs: []string{direct}, Format: fmtName, Headers: map[string]string{"Referer": urlHome}}
		if fmtName == "m3u8" {
			if rewritten, err := rewriteHoudaM3U8(c, direct, urlHome); err == nil {
				extra["m3u8_text"] = rewritten
			}
		}
	}
	if len(streams) == 0 {
		return nil, fmt.Errorf("houda lesson %s has no stream", lessonID)
	}
	return &extractor.MediaInfo{Site: "houda", Title: fmt.Sprintf("[%d]--%s", index, title), Streams: streams, Extra: extra}, nil
}

func resolveHoudaCSSL(c *util.Client, roomID, recordID, title string, headers map[string]string) (*shared.CssLcloudPlayInfo, error) {
	cc, err := resolveHoudaCCCallback(c, roomID, recordID, headers)
	if err != nil {
		return nil, err
	}
	liveRoomID := firstNonEmpty(cc.RoomID, roomID)
	accessID := firstNonEmpty(cc.UserID, cc.AccountID)
	viewerToken := firstNonEmpty(cc.ViewerToken, accessID+":"+liveRoomID)
	return shared.CssLcloudResolvePlayInfo(c, shared.CssLcloudPayload{
		LiveRoomID:  liveRoomID,
		UserID:      accessID,
		AccessID:    accessID,
		RecordID:    firstNonEmpty(cc.RecordID, recordID),
		ViewerName:  firstNonEmpty(cc.ViewerName, title),
		ViewerToken: viewerToken,
		Referer:     urlHome,
	})
}

type houdaCCInfo struct {
	UserID      string
	AccountID   string
	RoomID      string
	RecordID    string
	ViewerName  string
	ViewerToken string
}

func resolveHoudaCCCallback(c *util.Client, roomID, recordID string, headers map[string]string) (houdaCCInfo, error) {
	callbackURL := fmt.Sprintf(urlCCViewPlayback, url.PathEscape(roomID), url.PathEscape(recordID))
	finalURL, err := fetchRedirectLocation(c, callbackURL, headers)
	if err != nil {
		return houdaCCInfo{}, err
	}
	u, err := url.Parse(finalURL)
	if err != nil {
		return houdaCCInfo{}, fmt.Errorf("houda parse CSSL callback URL: %w", err)
	}
	q := u.Query()
	info := houdaCCInfo{
		UserID:      firstNonEmpty(q.Get("userId"), q.Get("userid"), q.Get("uid")),
		AccountID:   firstNonEmpty(q.Get("accountId"), q.Get("accessid"), q.Get("accessId")),
		RoomID:      firstNonEmpty(q.Get("roomId"), q.Get("roomid"), q.Get("room_id"), roomID),
		RecordID:    firstNonEmpty(q.Get("recordId"), q.Get("recordid"), q.Get("record_id"), recordID),
		ViewerName:  firstNonEmpty(q.Get("viewername"), q.Get("viewerName"), q.Get("userName")),
		ViewerToken: firstNonEmpty(q.Get("viewertoken"), q.Get("viewerToken"), q.Get("userToken")),
	}
	if info.AccountID == "" {
		info.AccountID = info.UserID
	}
	if info.UserID == "" || info.RoomID == "" || info.RecordID == "" {
		return houdaCCInfo{}, fmt.Errorf("houda CSSL callback missing userId/roomId/recordId: %s", finalURL)
	}
	return info, nil
}

func fetchRedirectLocation(c *util.Client, raw string, headers map[string]string) (string, error) {
	resp, err := c.Get(raw, headers)
	if err != nil {
		return "", fmt.Errorf("houda CSSL callback: %w", err)
	}
	defer resp.Body.Close()
	if resp.Request != nil && resp.Request.URL != nil {
		return resp.Request.URL.String(), nil
	}
	return raw, nil
}

func rewriteHoudaM3U8(c *util.Client, m3u8URL, referer string) (string, error) {
	body, err := c.GetString(m3u8URL, map[string]string{"Referer": referer})
	if err != nil {
		return "", err
	}
	return shared.CssLcloudRewriteM3U8Keys(c, body, referer)
}

func (s *Houda) extractCsslcloudURL(c *util.Client, rawURL string, headers map[string]string) (*extractor.MediaInfo, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	roomID := firstNonEmpty(q.Get("roomId"), q.Get("roomid"), q.Get("liveRoomId"), q.Get("room_id"))
	recordID := firstNonEmpty(q.Get("recordId"), q.Get("recordid"), q.Get("record_id"))
	accessID := firstNonEmpty(q.Get("userId"), q.Get("userid"), q.Get("accountId"), q.Get("accessid"))
	viewerToken := firstNonEmpty(q.Get("viewertoken"), q.Get("viewerToken"), accessID+":"+roomID)
	info, err := shared.CssLcloudResolvePlayInfo(c, shared.CssLcloudPayload{LiveRoomID: roomID, UserID: accessID, AccessID: accessID, RecordID: recordID, ViewerName: "houda", ViewerToken: viewerToken, Referer: urlHome})
	if err != nil {
		return nil, err
	}
	stream := extractor.Stream{Quality: "best", URLs: []string{info.VideoURL}, Format: mediaExt(info.VideoURL), AudioURL: info.AudioURL, Headers: map[string]string{"Referer": urlHome}}
	return &extractor.MediaInfo{Site: "houda", Title: "houda_csslcloud", Streams: map[string]extractor.Stream{"best": stream}}, nil
}

func parseClassID(rawURL string) string {
	if m := classIDRe.FindStringSubmatch(rawURL); len(m) > 1 {
		for i := 1; i < len(m); i++ {
			if m[i] != "" {
				return m[i]
			}
		}
	}
	return ""
}

func isCsslcloudURL(rawURL string) bool { return strings.Contains(rawURL, "csslcloud.net") }

func mediaExt(u string) string {
	lu := strings.ToLower(u)
	switch {
	case strings.Contains(lu, ".m3u8"):
		return "m3u8"
	case strings.Contains(lu, ".mp3"):
		return "mp3"
	default:
		return "mp4"
	}
}

func firstText(values ...any) string {
	for _, v := range values {
		if s := stringValue(v); s != "" {
			return s
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func stringValue(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(x)
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case json.Number:
		return x.String()
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}

func parseHoudaLessons(root map[string]any) []houdaLesson {
	data := unwrapHoudaData(root)
	rows := houdaMapList(data, "liveList", "list", "items", "records", "rows", "data")
	out := make([]houdaLesson, 0, len(rows))
	for _, row := range rows {
		lesson := houdaLessonFromAny(row)
		if firstText(lesson.ID, lesson.Title, lesson.Name, lesson.RoomID, lesson.RecordID, lesson.PlaybackURL, lesson.PlaybackMP4) == "" {
			continue
		}
		out = append(out, lesson)
	}
	return out
}

func unwrapHoudaData(v any) map[string]any {
	switch x := v.(type) {
	case map[string]any:
		if child, ok := x["data"]; ok {
			switch c := child.(type) {
			case map[string]any:
				return c
			case []any:
				return map[string]any{"list": c}
			}
		}
		return x
	case []any:
		return map[string]any{"list": x}
	default:
		return map[string]any{}
	}
}

func houdaMapList(v any, keys ...string) []map[string]any {
	var root any = v
	if m, ok := v.(map[string]any); ok {
		root = unwrapHoudaData(m)
	}
	switch x := root.(type) {
	case []any:
		return houdaMapsFromAnyList(x)
	case map[string]any:
		for _, key := range keys {
			if child, ok := x[key]; ok {
				if rows := houdaMapList(child); len(rows) > 0 {
					return rows
				}
			}
		}
		for _, key := range []string{"liveList", "lawList", "stageList", "dataList", "list", "items", "rows", "records", "data"} {
			if child, ok := x[key]; ok {
				if rows := houdaMapList(child); len(rows) > 0 {
					return rows
				}
			}
		}
	}
	return nil
}

func houdaMapsFromAnyList(values []any) []map[string]any {
	out := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if m, ok := value.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func firstMapText(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			if s := stringValue(value); s != "" {
				return s
			}
		}
	}
	return ""
}

func hasAnyMap(m map[string]any, keys ...string) bool {
	for _, key := range keys {
		if _, ok := m[key]; ok {
			return true
		}
	}
	return false
}

func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func appendStringUnique(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func collectHoudaLawRefs(raw map[string]any) []houdaLawRef {
	var out []houdaLawRef
	seen := map[string]bool{}
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			if id := firstMapText(x, "id", "lawId", "law_id"); id != "" && !seen[id] {
				seen[id] = true
				out = append(out, houdaLawRef{ID: id, Title: firstMapText(x, "name", "lawName", "title")})
			}
			for _, key := range []string{"lawList", "stageList", "children", "list", "items", "data"} {
				if child, ok := x[key]; ok {
					walk(child)
				}
			}
		case []any:
			for _, child := range x {
				walk(child)
			}
		}
	}
	walk(raw)
	return out
}

func houdaMaterialKey(item map[string]any) string {
	if id := firstMapText(item, "id", "materialId", "fileId"); id != "" {
		return "id:" + id
	}
	if u := normalizeHoudaURL(firstMapText(item, "downLoadUrl", "downloadUrl", "fileUrl", "url", "path")); u != "" {
		return "url:" + u
	}
	if title := firstMapText(item, "title", "name"); title != "" {
		return "title:" + title
	}
	return ""
}

func normalizeHoudaURL(raw string) string {
	raw = strings.TrimSpace(strings.Trim(raw, `"'`))
	raw = strings.ReplaceAll(raw, `\/`, `/`)
	switch {
	case raw == "":
		return ""
	case strings.HasPrefix(raw, "//"):
		return "http:" + raw
	case strings.HasPrefix(raw, "/"):
		u, err := url.Parse(urlOrigin)
		if err != nil {
			return raw
		}
		ref, err := url.Parse(raw)
		if err != nil {
			return raw
		}
		return u.ResolveReference(ref).String()
	default:
		return raw
	}
}

func houdaFileExt(hint, rawURL string) string {
	hint = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(hint)), ".")
	if hint != "" && hint != "file" {
		return hint
	}
	if u, err := url.Parse(rawURL); err == nil {
		if ext := strings.TrimPrefix(strings.ToLower(path.Ext(u.Path)), "."); ext != "" {
			return ext
		}
	}
	if ext := strings.TrimPrefix(strings.ToLower(path.Ext(rawURL)), "."); ext != "" {
		return ext
	}
	return "pdf"
}

func addHoudaQuery(api string, params map[string]string) string {
	u, err := url.Parse(api)
	if err != nil {
		return api
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func houdaLessonFromAny(v any) houdaLesson {
	m, ok := v.(map[string]any)
	if !ok {
		return houdaLesson{}
	}
	var lesson houdaLesson
	b, err := json.Marshal(m)
	if err == nil {
		_ = json.Unmarshal(b, &lesson)
	}
	if firstText(lesson.ID) == "" {
		lesson.ID = firstMapText(m, "id", "liveId", "lessonId")
	}
	return lesson
}

func mergeHoudaLesson(base, extra houdaLesson) houdaLesson {
	if firstText(base.ID) == "" {
		base.ID = extra.ID
	}
	if firstText(base.Title) == "" {
		base.Title = extra.Title
	}
	if firstText(base.Name) == "" {
		base.Name = extra.Name
	}
	if firstText(base.CourseName) == "" {
		base.CourseName = extra.CourseName
	}
	if firstText(base.CourseID) == "" {
		base.CourseID = extra.CourseID
	}
	if firstText(base.ClassID) == "" {
		base.ClassID = extra.ClassID
	}
	if firstText(base.Type) == "" {
		base.Type = extra.Type
	}
	if firstText(base.CCLiveID) == "" {
		base.CCLiveID = extra.CCLiveID
	}
	if firstText(base.RoomID) == "" {
		base.RoomID = extra.RoomID
	}
	if firstText(base.MainRoomID) == "" {
		base.MainRoomID = extra.MainRoomID
	}
	if firstText(base.RecordID) == "" {
		base.RecordID = extra.RecordID
	}
	if firstText(base.LiveURL) == "" {
		base.LiveURL = extra.LiveURL
	}
	if firstText(base.PlaybackURL) == "" {
		base.PlaybackURL = extra.PlaybackURL
	}
	if firstText(base.PlaybackMP4) == "" {
		base.PlaybackMP4 = extra.PlaybackMP4
	}
	if firstText(base.PlaybackMP3) == "" {
		base.PlaybackMP3 = extra.PlaybackMP3
	}
	if firstText(base.StageID) == "" {
		base.StageID = extra.StageID
	}
	if firstText(base.StageName) == "" {
		base.StageName = extra.StageName
	}
	if firstText(base.LawID) == "" {
		base.LawID = extra.LawID
	}
	if firstText(base.LawName) == "" {
		base.LawName = extra.LawName
	}
	return base
}
