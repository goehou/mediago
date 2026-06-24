package htknow

import (
	"net/url"
	"strings"
	"testing"
)

func TestAnswerEndpointConstantsMatchSource(t *testing.T) {
	want := map[string]string{
		"answerTagURL":         "https://saas.clientapi.htknow.com/pc_view/quest/get_quest_tag_list",
		"answerNumURL":         "https://saas.clientapi.htknow.com/pc_view/quest/get_quest_num_list",
		"answerListURL":        "https://saas.clientapi.htknow.com/pc_view/quest/get_quest_list",
		"answerCreatePaperURL": "https://saas.clientapi.htknow.com/pc_view/quest/create_question_paper",
	}
	got := map[string]string{
		"answerTagURL":         answerTagURL,
		"answerNumURL":         answerNumURL,
		"answerListURL":        answerListURL,
		"answerCreatePaperURL": answerCreatePaperURL,
	}
	for name, wantURL := range want {
		if got[name] != wantURL {
			t.Fatalf("%s = %q, want %q", name, got[name], wantURL)
		}
	}
}

func TestMediaFromSourcesKeepsHTMLOnlyEntry(t *testing.T) {
	html := "<h1>图文内容</h1>"
	mi, err := mediaFromSources("课程", []source{{name: "图文章节", kind: "图文", html: html}})
	if err != nil {
		t.Fatalf("mediaFromSources returned error: %v", err)
	}
	if mi.Title != "图文章节" {
		t.Fatalf("title = %q, want %q", mi.Title, "图文章节")
	}
	stream, ok := mi.Streams["document"]
	if !ok {
		t.Fatalf("document stream missing: %#v", mi.Streams)
	}
	if stream.Format != "html" {
		t.Fatalf("format = %q, want html", stream.Format)
	}
	if len(stream.URLs) != 1 || !strings.HasPrefix(stream.URLs[0], "data:text/html;charset=utf-8,") {
		t.Fatalf("document URL = %#v, want data:text/html", stream.URLs)
	}
	escaped := strings.TrimPrefix(stream.URLs[0], "data:text/html;charset=utf-8,")
	decoded, err := url.PathUnescape(escaped)
	if err != nil {
		t.Fatalf("decode html data URL: %v", err)
	}
	if decoded != html {
		t.Fatalf("decoded html = %q, want %q", decoded, html)
	}
	if mi.Extra["html_content"] != html {
		t.Fatalf("html_content extra = %#v, want %q", mi.Extra["html_content"], html)
	}
}

func TestMediaFromSourcesKeepsMixedVideoAndHTML(t *testing.T) {
	mi, err := mediaFromSources("课程", []source{
		{name: "视频", kind: "视频", url: "https://example.com/video.mp4"},
		{name: "图文", kind: "图文", html: "<p>content</p>"},
	})
	if err != nil {
		t.Fatalf("mediaFromSources returned error: %v", err)
	}
	if len(mi.Entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(mi.Entries))
	}
	if _, ok := mi.Entries[0].Streams["default"]; !ok {
		t.Fatalf("video default stream missing: %#v", mi.Entries[0].Streams)
	}
	if _, ok := mi.Entries[1].Streams["document"]; !ok {
		t.Fatalf("html document stream missing: %#v", mi.Entries[1].Streams)
	}
}
