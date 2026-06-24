package jianshe99

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/nichuanfang/medigo/internal/extractor"
	"github.com/nichuanfang/medigo/internal/extractor/shared"
	"github.com/nichuanfang/medigo/internal/util"
)

const (
	MEMBER_HOME_URL       = "https://member.jianshe99.com/homes/mycourse"
	ELEARNING_HOME_URL    = "https://elearning.jianshe99.com/"
	DOORMAN_BASE_URL      = "https://gateway.jianshe99.com/doorman/op/"
	MATERIALS_URL         = "https://elearning.jianshe99.com/xcware/myhome/teachingMaterials.shtm?cwareID=%s&identity=%s"
	material_download_url = "https://elearning.jianshe99.com/data2file/downloadFile/getVideoJyFileDocx?fileUrl=%s&fileName=%s"

	course_group_path    = "~/c-home/w-home/f/ru/userCourseClassList"
	course_detail_path   = "~/c-home/w-home/f/ru/getUserHomeCourse"
	course_subject_path  = "~/c-home/a-home/f/ru/getUserHomeCourse"
	live_replay_referer  = "https://live.cdeledu.com/"
	live_replay_info_url = "https://live.cdeledu.com/liveapi/entry/getReplayInfo"
	cc_replay_login_url  = "https://view.csslcloud.net/api/room/replay/login"
	cc_replay_vod_url    = "https://view.csslcloud.net/api/record/vod"
	cc_replay_version    = "3.6.1"
	video_list_url       = "https://elearning.jianshe99.com/xcware/video/videoList/videoList.shtm?cwareID=%s&courseIds=%s"
)

var patterns = []string{`(?:[\w-]+\.)?jianshe99\.com/|(?:[\w-]+\.)?cdeledu\.com/`}

func init() {
	extractor.Register(&Jianshe99{}, extractor.SiteInfo{Name: "Jianshe99", URL: "jianshe99.com", NeedAuth: true})
}

type Jianshe99 struct{}

func (j *Jianshe99) Patterns() []string { return patterns }

var (
	cwareRe        = regexp.MustCompile(`(?i)cwareID=([^&#]+)`)
	courseIDsRe    = regexp.MustCompile(`(?i)(?:courseIds|courseId)=([^&#]+)`)
	anchorRe       = regexp.MustCompile(`(?is)<a\b[^>]*(?:href|onclick)=["'][^"']*(?:videoPlay|/dispatch/th/live/callback/play)[^"']*["'][^>]*>.*?</a>`)
	urlInAttrRe    = regexp.MustCompile(`(?is)(https?:)?//[^"'<> ]+/dispatch/th/live/callback/play[^"'<> ]*|/dispatch/th/live/callback/play[^"'<> ]*`)
	onclickArgRe   = regexp.MustCompile(`(?is)(?:window\.open|videoPlay)\(\s*["']([^"']+)["']`)
	htmlTitleRe    = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	fieldStringRe  = regexp.MustCompile(`(?is)["']?%s["']?\s*[:=]\s*["']([^"']+)["']`)
	m3u8URLPattern = regexp.MustCompile(`(?i)\.m3u8(?:[?#].*)?$`)
)

func (j *Jianshe99) Extract(rawURL string, opts *extractor.ExtractOpts) (*extractor.MediaInfo, error) {
	if opts == nil || opts.Cookies == nil {
		return nil, fmt.Errorf("jianshe99 requires login cookies (use --cookies or --cookies-from-browser)")
	}
	c := util.NewClient()
	c.SetCookieJar(opts.Cookies)
	headers := map[string]string{"Referer": MEMBER_HOME_URL, "Origin": "https://member.jianshe99.com"}

	if isLiveReplayURL(rawURL) {
		entry, err := resolveReplay(c, rawURL, rawURL, "jianshe99_replay", nil)
		if err != nil {
			return nil, err
		}
		return entry, nil
	}

	pageURL := buildVideoListURL(rawURL)
	if pageURL == "" {
		return nil, fmt.Errorf("cannot parse jianshe99 cwareID from URL: %s", rawURL)
	}
	body, err := c.GetString(pageURL, headers)
	if err != nil {
		return nil, fmt.Errorf("fetch jianshe99 video list: %w", err)
	}
	lessons := parseLessons(body)
	if len(lessons) == 0 {
		return nil, fmt.Errorf("jianshe99: no live replay lesson URLs found")
	}

	entries := make([]*extractor.MediaInfo, 0, len(lessons))
	for i, lesson := range lessons {
		entry, err := resolveReplay(c, lesson.PlayURL, pageURL, firstNonEmpty(lesson.Title, fmt.Sprintf("课时%d", i+1)), []byte(body))
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("jianshe99: parsed lessons but no csslcloud media resolved")
	}
	return &extractor.MediaInfo{Site: "jianshe99", Title: firstNonEmpty(extractHTMLTitle(body), "jianshe99"), Entries: entries}, nil
}

type lessonRef struct {
	PlayURL string
	Title   string
}

func parseLessons(body string) []lessonRef {
	seen := map[string]bool{}
	var out []lessonRef
	for _, m := range anchorRe.FindAllString(body, -1) {
		u := firstNonEmpty(extractFirst(onclickArgRe, m), extractFirst(urlInAttrRe, m))
		u = normalizeURL(html.UnescapeString(u))
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, lessonRef{PlayURL: u, Title: cleanText(stripTags(m))})
	}
	if len(out) == 0 {
		for _, u := range urlInAttrRe.FindAllString(body, -1) {
			u = normalizeURL(html.UnescapeString(u))
			if u != "" && !seen[u] {
				seen[u] = true
				out = append(out, lessonRef{PlayURL: u})
			}
		}
	}
	return out
}

func resolveReplay(c *util.Client, playURL, referer, title string, pageBody []byte) (*extractor.MediaInfo, error) {
	payload, err := fetchReplayPayload(c, playURL, referer)
	if err != nil && len(pageBody) > 0 {
		payload = payloadFromText(string(pageBody))
	}
	if err != nil && payload.LiveRoomID == "" {
		payload = payloadFromQuery(playURL)
	}
	if payload.LiveRoomID == "" || payload.AccessID == "" || payload.RecordID == "" {
		return nil, fmt.Errorf("jianshe99 replay payload missing liveRoomId/accessid/recordId")
	}
	payload.Referer = live_replay_referer
	playInfo, err := shared.CssLcloudResolvePlayInfo(c, payload)
	if err != nil {
		return nil, err
	}

	extra := map[string]any{"session_id": playInfo.SessionID, "record_id": payload.RecordID}
	if m3u8URLPattern.MatchString(playInfo.VideoURL) {
		text, err := c.GetString(playInfo.VideoURL, map[string]string{"Referer": live_replay_referer})
		if err == nil {
			if prepared, err := shared.CssLcloudRewriteM3U8Keys(c, text, live_replay_referer); err == nil {
				extra["prepared_m3u8_text"] = prepared
			}
		}
	}
	return &extractor.MediaInfo{
		Site:  "jianshe99",
		Title: util.SanitizeFilename(firstNonEmpty(title, payload.RecordID)),
		Streams: map[string]extractor.Stream{
			"best": {
				Quality:  "best",
				URLs:     []string{playInfo.VideoURL},
				Format:   pickFormat(playInfo.VideoURL),
				AudioURL: playInfo.AudioURL,
				Headers:  map[string]string{"Referer": live_replay_referer},
			},
		},
		Extra: extra,
	}, nil
}

type replayInfoResponse struct {
	Result any `json:"result"`
	Data   struct {
		Replay replayPayload `json:"replay"`
		Vod    replayPayload `json:"vod"`
	} `json:"data"`
	Replay replayPayload `json:"replay"`
	Vod    replayPayload `json:"vod"`
}

type replayPayload struct {
	LiveRoomID  string `json:"liveRoomId"`
	LiveID      string `json:"liveId"`
	RoomID      string `json:"roomid"`
	AccessID    string `json:"accessid"`
	AccessKey   string `json:"accesskey"`
	RecordID    string `json:"recordId"`
	RecordIDAlt string `json:"recordid"`
	UserID      string `json:"userid"`
	UID         string `json:"uid"`
	ViewerName  string `json:"viewername"`
	ViewerToken string `json:"viewertoken"`
	Token       string `json:"token"`
}

func fetchReplayPayload(c *util.Client, playURL, referer string) (shared.CssLcloudPayload, error) {
	api := live_replay_info_url + "?url=" + url.QueryEscape(stripFragment(playURL))
	body, err := c.GetString(api, map[string]string{"Referer": firstNonEmpty(referer, live_replay_referer)})
	if err != nil {
		return shared.CssLcloudPayload{}, err
	}
	var resp replayInfoResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return shared.CssLcloudPayload{}, err
	}
	candidates := []replayPayload{resp.Data.Replay, resp.Data.Vod, resp.Replay, resp.Vod}
	for _, p := range candidates {
		if out := normalizePayload(p); out.LiveRoomID != "" || out.RecordID != "" {
			return out, nil
		}
	}
	return payloadFromText(body), nil
}

func normalizePayload(p replayPayload) shared.CssLcloudPayload {
	liveID := firstNonEmpty(p.LiveRoomID, p.LiveID, p.RoomID)
	userID := firstNonEmpty(p.UserID, p.UID)
	token := firstNonEmpty(p.ViewerToken, p.Token)
	return shared.CssLcloudPayload{
		LiveRoomID:  liveID,
		UserID:      userID,
		AccessID:    firstNonEmpty(p.AccessID, p.AccessKey),
		RecordID:    firstNonEmpty(p.RecordID, p.RecordIDAlt),
		ViewerName:  firstNonEmpty(p.ViewerName, userID),
		ViewerToken: token,
	}
}

func payloadFromText(s string) shared.CssLcloudPayload {
	return shared.CssLcloudPayload{
		LiveRoomID:  firstField(s, "liveRoomId", "liveId", "roomid", "liveid"),
		UserID:      firstField(s, "userid", "uid", "userId"),
		AccessID:    firstField(s, "accessid", "accessId", "accesskey"),
		RecordID:    firstField(s, "recordId", "recordid"),
		ViewerName:  firstField(s, "viewername", "viewerName"),
		ViewerToken: firstField(s, "viewertoken", "viewerToken", "token"),
	}
}

func payloadFromQuery(raw string) shared.CssLcloudPayload {
	u, err := url.Parse(raw)
	if err != nil {
		return shared.CssLcloudPayload{}
	}
	q := u.Query()
	return shared.CssLcloudPayload{
		LiveRoomID:  firstNonEmpty(q.Get("liveRoomId"), q.Get("liveId"), q.Get("roomid"), q.Get("liveid")),
		UserID:      firstNonEmpty(q.Get("userid"), q.Get("uid")),
		AccessID:    firstNonEmpty(q.Get("accessid"), q.Get("accessId"), q.Get("accesskey")),
		RecordID:    firstNonEmpty(q.Get("recordId"), q.Get("recordid")),
		ViewerName:  firstNonEmpty(q.Get("viewername"), q.Get("userid")),
		ViewerToken: firstNonEmpty(q.Get("viewertoken"), q.Get("viewerToken"), q.Get("token")),
	}
}

func buildVideoListURL(raw string) string {
	if strings.Contains(raw, "videoList.shtm") {
		return raw
	}
	cwareID := extractFirst(cwareRe, raw)
	if cwareID == "" {
		return ""
	}
	courseIDs := extractFirst(courseIDsRe, raw)
	return fmt.Sprintf(video_list_url, url.QueryEscape(cwareID), url.QueryEscape(courseIDs))
}

func firstField(s string, names ...string) string {
	for _, name := range names {
		re := regexp.MustCompile(fmt.Sprintf(fieldStringRe.String(), regexp.QuoteMeta(name)))
		if v := extractFirst(re, s); v != "" {
			return html.UnescapeString(v)
		}
	}
	return ""
}

func isLiveReplayURL(s string) bool { return strings.Contains(s, "/dispatch/th/live/callback/play") }

func stripFragment(s string) string {
	if i := strings.IndexByte(s, '#'); i >= 0 {
		return s[:i]
	}
	return s
}

func normalizeURL(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, `\/`, `/`))
	if strings.HasPrefix(s, "//") {
		return "https:" + s
	}
	if strings.HasPrefix(s, "/") {
		return strings.TrimRight(ELEARNING_HOME_URL, "/") + s
	}
	return s
}

func extractHTMLTitle(body string) string {
	return cleanText(stripTags(extractFirst(htmlTitleRe, body)))
}

func extractFirst(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	for _, g := range m[1:] {
		if g != "" {
			return g
		}
	}
	return ""
}

func cleanText(s string) string { return strings.Join(strings.Fields(html.UnescapeString(s)), " ") }

func stripTags(s string) string {
	return regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(s, " ")
}

func pickFormat(s string) string {
	if strings.Contains(strings.ToLower(s), ".m3u8") {
		return "m3u8"
	}
	return "mp4"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
