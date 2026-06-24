package med66

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func collectMaps(v any) []anyMap {
	var out []anyMap
	var walk func(any)
	walk = func(x any) {
		switch t := x.(type) {
		case map[string]any:
			out = append(out, anyMap(t))
			for _, child := range t {
				walk(child)
			}
		case []any:
			for _, child := range t {
				walk(child)
			}
		}
	}
	walk(v)
	return out
}

func firstString(m anyMap, keys ...string) string {
	for _, k := range keys {
		if s := strings.TrimSpace(strAny(m[k])); s != "" && s != "<nil>" {
			return s
		}
	}
	return ""
}

func strAny(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

func normalizeURL(s, base string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "//") {
		return "https:" + s
	}
	if strings.HasPrefix(s, "/") {
		b, _ := url.Parse(base)
		u, _ := url.Parse(s)
		return b.ResolveReference(u).String()
	}
	return s
}

func pickFormat(u string) string {
	if strings.Contains(u, ".m3u8") {
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

func uniqueNonEmpty(values ...string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func cookieValue(jar http.CookieJar, bases []string, name string) string {
	for _, raw := range bases {
		u, _ := url.Parse(raw)
		for _, c := range jar.Cookies(u) {
			if c.Name == name {
				return c.Value
			}
		}
	}
	return ""
}
