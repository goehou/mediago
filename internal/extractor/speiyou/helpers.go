package speiyou

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func authFromJar(j http.CookieJar) authInfo {
	var a authInfo
	var parts []string
	for _, host := range []string{referer, "https://course-api-online.speiyou.com/", "https://classroom-api-online.speiyou.com/"} {
		u, _ := url.Parse(host)
		for _, ck := range j.Cookies(u) {
			parts = append(parts, ck.Name+"="+ck.Value)
			if tokenName(ck.Name) {
				a.Token = first(a.Token, ck.Value)
			}
			if strings.EqualFold(ck.Name, "stuId") || strings.EqualFold(ck.Name, "stu_id") || strings.EqualFold(ck.Name, "pu_uid") {
				a.StuID = first(a.StuID, ck.Value)
			}
			if strings.HasPrefix(strings.TrimSpace(ck.Value), "{") {
				var m map[string]any
				if json.Unmarshal([]byte(ck.Value), &m) == nil {
					a.Token = first(a.Token, findText(m, "token", "hb_token", "passport_token", "signToken"))
					a.StuID = first(a.StuID, findText(m, "stuId", "stu_id", "pu_uid", "puUid", "studentId", "student_id"))
				}
			}
		}
	}
	a.Cookie = strings.Join(parts, "; ")
	return a
}
func baseHeaders(a authInfo) map[string]string {
	return map[string]string{"User-Agent": USER_AGENT, "referer": referer, "resVer": "1.0.6", "version": "3.60.0.2368", "terminal": "pc", "lang": "ch", "appClientType": "xes", "Referer": referer, "Origin": "owcr://classroom", "Accept": "application/json, text/plain, */*", "token": a.Token, "authorization": a.Token, "cookie": a.Cookie, "Cookie": a.Cookie, "stuId": a.StuID}
}
func tokenName(n string) bool {
	n = strings.ToLower(n)
	return n == "token" || n == "hb_token" || n == "passport_token" || n == "signtoken" || n == "authorization"
}
func courseKey(m map[string]any) string {
	return first(textAt(m, "stdCourseId", "course_id", "courseId"), textAt(unwrapMap(m["courseInfo"]), "stdCourseId", "courseId"))
}
func lessonKey(m map[string]any) string {
	return first(textAt(m, "liveId", "live_id"), textAt(unwrapMap(m["liveInfo"]), "liveId"))
}
func jsonToMaps(v any) []map[string]any {
	if l, ok := v.([]any); ok {
		return maps(l)
	}
	m := unwrapMap(v)
	for _, k := range []string{"data", "list", "result", "records"} {
		if l, ok := m[k].([]any); ok {
			return maps(l)
		}
		if mm, ok := m[k].(map[string]any); ok {
			if out := jsonToMaps(mm); len(out) > 0 {
				return out
			}
		}
	}
	return nil
}
func maps(in []any) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, v := range in {
		if m, ok := v.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}
func unwrapMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}
func valueStrings(v any) []string {
	switch x := v.(type) {
	case string:
		return []string{x}
	case []any:
		out := make([]string, 0, len(x))
		for _, e := range x {
			out = append(out, fmt.Sprint(e))
		}
		return out
	case []string:
		return x
	}
	return nil
}
func findURL(v any) string {
	switch t := v.(type) {
	case map[string]any:
		for _, k := range []string{"videoUrl", "url", "playUrl"} {
			if u := textAt(t, k); strings.HasPrefix(u, "http") {
				return u
			}
		}
		for _, x := range t {
			if u := findURL(x); u != "" {
				return u
			}
		}
	case []any:
		for _, x := range t {
			if u := findURL(x); u != "" {
				return u
			}
		}
	}
	return ""
}
func findText(v any, keys ...string) string {
	switch t := v.(type) {
	case map[string]any:
		if s := textAt(t, keys...); s != "" {
			return s
		}
		for _, x := range t {
			if s := findText(x, keys...); s != "" {
				return s
			}
		}
	case []any:
		for _, x := range t {
			if s := findText(x, keys...); s != "" {
				return s
			}
		}
	}
	return ""
}
func textAt(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && fmt.Sprint(v) != "<nil>" {
			return strings.TrimSpace(fmt.Sprint(v))
		}
	}
	return ""
}
func intAt(m map[string]any, key string) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}
func clone(h map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range h {
		out[k] = v
	}
	return out
}
func match1(s, pat string) string {
	if m := regexp.MustCompile(pat).FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(html.UnescapeString(m[1]))
	}
	return ""
}
func first(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" && strings.TrimSpace(v) != "<nil>" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
func sanitize(s string) string {
	s = html.UnescapeString(strings.TrimSpace(s))
	return regexp.MustCompile(`[\\/:*?"<>|\r\n\t]+`).ReplaceAllString(s, "_")
}
func pickFormat(u string) string {
	p := strings.ToLower(strings.SplitN(strings.SplitN(u, "?", 2)[0], "#", 2)[0])
	if i := strings.LastIndex(p, "."); i >= 0 && i < len(p)-1 {
		return p[i+1:]
	}
	return "mp4"
}
