package gaotu

import "testing"

func TestEndpointsForBrandDomains(t *testing.T) {
	tests := []struct {
		name      string
		rawURL    string
		courseURL string
		infoURL   string
		videoURL  string
		liveURL   string
		sourceURL string
		fileURL   string
		priceURL  string
		referer   string
	}{
		{
			name:      "gaotu",
			rawURL:    "https://www.gaotu.cn/course?clazzNumber=G001",
			courseURL: "https://api.gaotu.cn/studyPlatform/v1/unit/clazz/list?isDebounce=true&os=h5-pc&p_client=1",
			infoURL:   "https://interactive.gaotu.cn/live/api/studyCenter/v1/user/pc/clazz/detail",
			videoURL:  "https://api.gaotu.cn/live/zplan/login/videoLive",
			liveURL:   "https://interactive.gaotu.cn/live/api/live/zplan/playbackWeb",
			sourceURL: "https://interactive.gaotu.cn/live/api/pan/listDir",
			fileURL:   "https://interactive.gaotu.cn/live/api/pan/file",
			priceURL:  "https://api.gaotu.cn/cs/api/product/course/detailButton?productSpuNumber=%s",
			referer:   "https://www.gaotu.cn",
		},
		{
			name:      "tutu",
			rawURL:    "https://gaotu100.com/course?clazzNumber=T001",
			courseURL: "https://api.gaotu100.com/studyPlatform/v1/unit/clazz/list?isDebounce=true&os=h5-pc&p_client=2",
			infoURL:   "https://interactive.gaotu100.com/live/api/studyCenter/v1/user/pc/clazz/detail",
			videoURL:  "https://api.gaotu100.com/live/zplan/login/videoLive",
			liveURL:   "https://interactive.gaotu100.com/live/api/live/zplan/playbackWeb",
			sourceURL: "https://interactive.gaotu100.com/live/api/pan/listDir",
			fileURL:   "https://interactive.gaotu100.com/live/api/pan/file",
			priceURL:  "https://api.gaotu100.com/cs/api/product/course/detailButton?productSpuNumber=%s",
			referer:   "https://gaotu100.com",
		},
		{
			name:      "gaozhong",
			rawURL:    "https://www.gtgz.cn/course?clazzNumber=H001",
			courseURL: "https://api.gtgz.cn/studyPlatform/v1/unit/clazz/list?isDebounce=true&os=h5-pc&p_client=8",
			infoURL:   "https://interactive.gtgz.cn/live/api/studyCenter/v1/user/pc/clazz/detail",
			videoURL:  "https://api.gtgz.cn/live/zplan/login/videoLive",
			liveURL:   "https://interactive.gtgz.cn/live/api/live/zplan/playbackWeb",
			sourceURL: "https://interactive.gtgz.cn/live/api/pan/listDir",
			fileURL:   "https://interactive.gtgz.cn/live/api/pan/file",
			priceURL:  "https://api.gtgz.cn/cs/api/product/course/detailButton?productSpuNumber=%s",
			referer:   "https://www.gtgz.cn",
		},
		{
			name:      "suyang",
			rawURL:    "https://www.naiyouxuexi.com/course?clazzNumber=S001",
			courseURL: "https://api.naiyouxuexi.com/studyPlatform/v1/unit/clazz/list?isDebounce=true&os=h5-pc&p_client=18",
			infoURL:   "https://interactive.naiyouxuexi.com/live/api/studyCenter/v1/user/pc/clazz/detail",
			videoURL:  "https://api.naiyouxuexi.com/live/zplan/login/videoLive",
			liveURL:   "https://interactive.naiyouxuexi.com/live/api/live/zplan/playbackWeb",
			sourceURL: "https://interactive.naiyouxuexi.com/live/api/pan/listDir",
			fileURL:   "https://interactive.naiyouxuexi.com/live/api/pan/file",
			priceURL:  "https://api.naiyouxuexi.com/cs/api/product/course/detailButton?productSpuNumber=%s",
			referer:   "https://www.naiyouxuexi.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := endpointsFor(tt.rawURL)
			if got.referer != tt.referer {
				t.Fatalf("referer = %q, want %q", got.referer, tt.referer)
			}
			if got.courseURL() != tt.courseURL {
				t.Fatalf("courseURL = %q, want %q", got.courseURL(), tt.courseURL)
			}
			if got.infoURL() != tt.infoURL {
				t.Fatalf("infoURL = %q, want %q", got.infoURL(), tt.infoURL)
			}
			if got.videoURL() != tt.videoURL {
				t.Fatalf("videoURL = %q, want %q", got.videoURL(), tt.videoURL)
			}
			if got.liveURL() != tt.liveURL {
				t.Fatalf("liveURL = %q, want %q", got.liveURL(), tt.liveURL)
			}
			if got.sourceURL() != tt.sourceURL {
				t.Fatalf("sourceURL = %q, want %q", got.sourceURL(), tt.sourceURL)
			}
			if got.fileURL() != tt.fileURL {
				t.Fatalf("fileURL = %q, want %q", got.fileURL(), tt.fileURL)
			}
			if got.priceURL() != tt.priceURL {
				t.Fatalf("priceURL = %q, want %q", got.priceURL(), tt.priceURL)
			}
		})
	}
}
