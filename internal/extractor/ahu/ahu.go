// Package ahu implements an extractor for ahuyikao.com (安徽医考).
//
// API endpoints from decompiled Mooc/Courses/Ahu/Ahu_Course.pyc:
//
//	https://www.ahuyikao.com/                                                    (home)
//	https://www.ahuyikao.com/center/mycourse.html                                (purchased list)
//	https://www.ahuyikao.com/course/courseinfo.html?courseId={cid}               (course detail)
//	https://www.ahuyikao.com/video/videoplay.html?courseId={cid}&lessonId={lid}  (lesson play page)
//
// The course detail page is HTML; we extract title and lesson list from it.
package ahu

import (
	"fmt"
	"regexp"

	"github.com/nichuanfang/medigo/internal/extractor"
	"github.com/nichuanfang/medigo/internal/util"
)

const (
	urlHome       = "https://www.ahuyikao.com/"
	urlMyCourse   = "https://www.ahuyikao.com/center/mycourse.html"
	urlCourseInfo = "https://www.ahuyikao.com/course/courseinfo.html?courseId=%s"
	urlVideoPlay  = "https://www.ahuyikao.com/video/videoplay.html?courseId=%s&lessonId=%s"
)

var patterns = []string{`(?:[\w-]+\.)?ahuyikao\.com/`}

func init() {
	extractor.Register(&Ahu{}, extractor.SiteInfo{Name: "Ahu", URL: "ahuyikao.com", NeedAuth: true})
}

type Ahu struct{}

func (a *Ahu) Patterns() []string { return patterns }

var (
	cidRe       = regexp.MustCompile(`courseId=(\d+)`)
	titleRe     = regexp.MustCompile(`<title>(.*?)</title>`)
	lessonItemRe = regexp.MustCompile(`lessonId\s*=\s*["'](\d+)["'][\s\S]*?title\s*=\s*["']([^"']+)["']`)
)

func (a *Ahu) Extract(rawURL string, opts *extractor.ExtractOpts) (*extractor.MediaInfo, error) {
	if opts == nil || opts.Cookies == nil {
		return nil, fmt.Errorf("ahu requires login cookies")
	}
	m := cidRe.FindStringSubmatch(rawURL)
	if m == nil {
		return nil, fmt.Errorf("cannot parse courseId from URL: %s", rawURL)
	}
	cid := m[1]

	c := util.NewClient()
	c.SetCookieJar(opts.Cookies)
	headers := map[string]string{"Referer": urlHome}

	pageURL := fmt.Sprintf(urlCourseInfo, cid)
	body, err := c.GetString(pageURL, headers)
	if err != nil {
		return nil, fmt.Errorf("fetch course info: %w", err)
	}

	title := "ahu_" + cid
	if t := titleRe.FindStringSubmatch(body); len(t) > 1 {
		title = t[1]
	}

	var entries []*extractor.MediaInfo
	for _, m := range lessonItemRe.FindAllStringSubmatch(body, -1) {
		lessonID, lessonTitle := m[1], m[2]
		entries = append(entries, &extractor.MediaInfo{
			Site:  "ahu",
			Title: lessonTitle,
			Streams: map[string]extractor.Stream{
				"page": {
					Quality: "best",
					URLs:    []string{fmt.Sprintf(urlVideoPlay, cid, lessonID)},
					Format:  "page", // page URL — actual mp4 needs further extraction
					Headers: headers,
				},
			},
			Extra: map[string]any{"lesson_id": lessonID, "course_id": cid},
		})
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("ahu: no lessons found in course page (course locked or HTML schema changed)")
	}

	return &extractor.MediaInfo{
		Site:    "ahu",
		Title:   title,
		Entries: entries,
	}, nil
}
