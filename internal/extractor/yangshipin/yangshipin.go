package yangshipin

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Sophomoresty/mediago/internal/extractor"
	"github.com/Sophomoresty/mediago/internal/util"
)

//go:embed ysp_timeshift.py
var yspTimeshiftPy []byte

var channels = map[string]string{
	"cctv1": "600001859", "cctv2": "600001800", "cctv3": "600001801",
	"cctv4": "600001814", "cctv5": "600001818", "cctv5+": "600001826",
	"cctv6": "600001802", "cctv7": "600001803", "cctv8": "600001804",
	"cctv9": "600001805", "cctv10": "600001806", "cctv11": "600001807",
	"cctv12": "600001808", "cctv13": "600001811", "cctv14": "600001809",
	"cctv15": "600001815", "cctv16": "600002188", "cctv17": "600001810",
}

var channelNames = map[string]string{
	"600001859": "CCTV-1 综合", "600001800": "CCTV-2 财经",
	"600001801": "CCTV-3 综艺", "600001814": "CCTV-4 中文国际",
	"600001818": "CCTV-5 体育", "600001826": "CCTV-5+ 体育赛事",
	"600001802": "CCTV-6 电影", "600001803": "CCTV-7 国防军事",
	"600001804": "CCTV-8 电视剧", "600001805": "CCTV-9 纪录",
	"600001806": "CCTV-10 科教", "600001807": "CCTV-11 戏曲",
	"600001808": "CCTV-12 社会与法", "600001811": "CCTV-13 新闻",
	"600001809": "CCTV-14 少儿", "600001815": "CCTV-15 音乐",
	"600002188": "CCTV-16 奥林匹克", "600001810": "CCTV-17 农业农村",
}

var patterns = []string{
	`yangshipin\.cn/tv/home`,
	`cctv\d+(?:\+)?-live`,
}

func init() {
	extractor.Register(&YSP{}, extractor.SiteInfo{
		Name: "yangshipin",
		URL:  "www.yangshipin.cn",
	})
}

type YSP struct{}

func (y *YSP) Patterns() []string { return patterns }

func (y *YSP) Extract(rawURL string, opts *extractor.ExtractOpts) (*extractor.MediaInfo, error) {
	pid, start, end, duration := parseInput(rawURL)
	if pid == "" {
		return nil, fmt.Errorf("cannot determine channel from %q (use cctv1-live, cctv13-live, or yangshipin.cn/tv/home?pid=...)", rawURL)
	}

	title := channelNames[pid]
	if title == "" {
		title = "YSP " + pid
	}

	now := time.Now().Unix()
	if start == 0 && end == 0 {
		start = now - 300
		end = now
		title += " 直播"
	} else {
		if end == 0 {
			end = start + int64(duration)
		}
		title += fmt.Sprintf(" 回放 %s", time.Unix(start, 0).Format("2006-01-02 15:04"))
	}

	m3u8URL, err := callTimeshift(pid, start, end)
	if err != nil {
		return nil, fmt.Errorf("timeshift API: %w", err)
	}

	return &extractor.MediaInfo{
		Site:  "yangshipin",
		Title: title,
		Streams: map[string]extractor.Stream{
			"hls": {
				Quality: "1080p",
				URLs:    []string{m3u8URL},
				Format:  "m3u8",
				Headers: map[string]string{
					"User-Agent": util.RandomUA(),
				},
			},
		},
	}, nil
}

func parseInput(raw string) (pid string, start, end int64, duration int) {
	raw = strings.TrimSpace(raw)

	re := regexp.MustCompile(`(?i)(cctv\d+\+?)-live`)
	if m := re.FindStringSubmatch(raw); len(m) > 1 {
		key := strings.ToLower(m[1])
		pid = channels[key]
		return
	}

	if strings.Contains(raw, "yangshipin.cn") {
		u, err := url.Parse(raw)
		if err == nil {
			if p := u.Query().Get("pid"); p != "" {
				pid = p
			}
			if s := u.Query().Get("start"); s != "" {
				start, _ = strconv.ParseInt(s, 10, 64)
			}
			if e := u.Query().Get("end"); e != "" {
				end, _ = strconv.ParseInt(e, 10, 64)
			}
			if d := u.Query().Get("duration"); d != "" {
				duration, _ = strconv.Atoi(d)
			}
		}
	}

	if pid == "" {
		key := strings.ToLower(strings.TrimSpace(raw))
		pid = channels[key]
	}
	return
}

func callTimeshift(pid string, start, end int64) (string, error) {
	python, err := findPython()
	if err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", "mediago-ysp-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	scriptPath := filepath.Join(tmpDir, "ysp_timeshift.py")
	if err := os.WriteFile(scriptPath, yspTimeshiftPy, 0o644); err != nil {
		return "", err
	}

	cmd := exec.Command(python, scriptPath,
		"--pid", pid,
		"--sid", "2024078201",
		"--stream", "fhd",
		"--start", strconv.FormatInt(start, 10),
		"--end", strconv.FormatInt(end, 10),
	)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("timeshift script: %s", ee.Stderr)
		}
		return "", err
	}

	var result struct {
		OK    bool   `json:"ok"`
		M3U8  string `json:"m3u8"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return "", fmt.Errorf("bad timeshift response: %w", err)
	}
	if !result.OK {
		return "", fmt.Errorf("timeshift: %s", result.Error)
	}
	if result.M3U8 == "" {
		return "", fmt.Errorf("timeshift returned empty m3u8")
	}
	return result.M3U8, nil
}

func findPython() (string, error) {
	for _, name := range []string{"python3", "python"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("python3 not found in PATH")
}
