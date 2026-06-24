# worker-2 full extractor audit

Scope: `duanshu enetedu eoffcn feishu fenbi gaodun gaotu gongxuanwang haiyangknow haozaixian houda houdu hqwx htknow huatu`.

Checks performed:
- Compared `internal/extractor/<site>/<site>.go` and sibling Go files with decompiled `.cdc.py` under `~/code/xwz-downloader-source-release/decompiled_full/Mooc/Courses/<Site>/`.
- Checked URL constants, HTTP method, JSON key/tag coverage, auth header/cookie flow, nil/error/body-close/dead-code risks.
- Baseline `python3 scripts/verify_full_alignment.py` reports all 15 scoped sites as `PASS`, but the issues below are manual source-alignment/code-review findings not caught by that script.

## duanshu

Issues:
- Auth/header mismatch: source `Duanshu_Base.py:61-68` initializes `user-agent`, `referer`, `x-member`, `x-shop`, `x-platform`, `X-CLIENT-VERSION`, `HTTP-X-H5-VERSION`; Go `duanshuHeaders` only emits `Accept`, `referer`/`Referer`, `x-member`, `x-shop` (`internal/extractor/duanshu/duanshu.go:269-287`). This is not byte-for-byte auth/header alignment.
- Unchecked auth probe error: source `_verify_login` calls `/h5/user/detail` as login verification (`Duanshu_Base.py:400-405`), but Go ignores the result with `_, _ = requestJSON(c, user_detail_url, headers)` (`internal/extractor/duanshu/duanshu.go:77`). A bad/expired cookie can be masked until later unrelated calls fail.
- No response-body leak found in direct HTTP paths reviewed.

## enetedu

NO ISSUE.

Evidence:
- URL constants and method map in `internal/extractor/enetedu/SOURCE_ALIGN.md` match source paths from `Enetedu_Base.py` / `Enetedu_Course.py`.
- Direct POST helper closes response bodies; no nil panic, unchecked material error, unused import, or dead code found in reviewed package paths.

## eoffcn

Issues:
- Unchecked public-key request: source defines and uses `pub_key_url` / `encrypt_url` (`Eoffcn_Course.py:47-48`) for the watch-demand flow; Go calls `c.GetString(pub_key_url, headers)` but discards both response and error before posting to `encrypt_url` (`internal/extractor/eoffcn/eoffcn.go:197-200`). If public-key/bootstrap fails, the later empty-string fallback hides the real failure cause.
- URL constants, GET/POST split, and JSON media key coverage otherwise align with `internal/extractor/eoffcn/SOURCE_ALIGN.md`.
- No response-body leak found.

## feishu

Issues:
- CRITICAL: unsupported source branches remain in the extractor. Source handles `/wiki`, `/docx`/`/docs`, `/file`, and `/minutes` (`Feishu_Course.py:354-386`), but Go only resolves `/minutes`; `/file` and `/docx`/`/wiki` return hard errors (`internal/extractor/feishu/feishu.go:62-68`). This is a source-alignment miss even though the broad verifier marks the site as PASS.
- Missing self-review artifact: `internal/extractor/feishu/SOURCE_ALIGN.md` does not exist, unlike the other reviewed sites.
- The implemented `/minutes` path uses source referer `https://www.feishu.cn` (`Feishu_Course.py:41`; `internal/extractor/feishu/feishu.go:60`) and no direct response-body leak was found.

## fenbi

Issues:
- Unchecked login validation: source `_check_cookie` probes both `login_check_url` and `ke_check_url` and sets login state only after successful responses (`Fenbi_Base.py:31-32`, `Fenbi_Base.py:199-237`), but Go discards `checkLogin` error (`internal/extractor/fenbi/fenbi.go:85`). This can mask expired cookies and defer the failure to unrelated course/media requests.
- URL constants, route parsing, GET methods, and media-meta JSON key coverage otherwise align with `internal/extractor/fenbi/SOURCE_ALIGN.md`.
- No response-body leak found.

## gaodun

NO ISSUE.

Evidence:
- URL constants from `Gaodun_Course.py:37-51` are represented in Go and documented in `internal/extractor/gaodun/SOURCE_ALIGN.md`.
- Reviewed direct request paths check errors and close bodies through shared helpers; no nil panic, unused import, or dead code found.

## gaotu

Issues:
- CRITICAL: registered URL patterns include `gaotu100.com`, `gtgz.cn`, and `naiyouxuexi.com` (`internal/extractor/gaotu/gaotu.go:27`), but Go constants always call `api.gaotu.cn` / `interactive.gaotu.cn` (`internal/extractor/gaotu/gaotu.go:16-24`). Source subclasses use brand-specific hosts and `p_client` values: `api.gaotu100.com`/`interactive.gaotu100.com` (`Gaotu_Tutu.py:40-48`), `api.gtgz.cn`/`interactive.gtgz.cn` (`Gaotu_Gaozhong.py:40-48`), and `api.naiyouxuexi.com`/`interactive.naiyouxuexi.com` (`Gaotu_Suyang.py:40-48`). Go only swaps Referer (`internal/extractor/gaotu/helpers.go:125-135`), so non-`gaotu.cn` domains are sent to fabricated/wrong API hosts.
- Source `price_url`, `source_url`, and `file_url` are part of the base flow (`Gaotu_Course.py:43-45`, `_get_price` at `Gaotu_Course.py:103-120`), but Go only declares these constants and has no call sites for `source_url`, `file_url`, or `price_url`. `internal/extractor/gaotu/SOURCE_ALIGN.md` marks them as aligned, but runtime coverage is missing.
- No response-body leak found in direct POST paths.

## gongxuanwang

Issues:
- Unchecked system-detail errors: `getSystemInfos` ignores errors from `loadPagedPostRows(system_course_detail_api, ...)` on both primary and alternate payloads (`internal/extractor/gongxuanwang/infos.go:147-150`). This can turn an HTTP/JSON failure into a later generic `empty rows` error and loses the source-equivalent failure reason.
- URL constants, auth token headers, Polyv delegation, and JSON field coverage otherwise align with `internal/extractor/gongxuanwang/SOURCE_ALIGN.md`.
- Direct POST helper closes response bodies; no nil panic found.

## haiyangknow

Issues:
- Unchecked JSON parse in Aliyun license flow: Go ignores `json.Unmarshal` failure for the license response (`internal/extractor/haiyangknow/aliyun.go:236`), then proceeds with an empty `root` map. This is a code-review issue because malformed license responses lose their parse error and fail later with less precise state.
- Low-risk unchecked entropy read: `rand.Read` return values are ignored when generating nonce bytes (`internal/extractor/haiyangknow/aliyun.go:281`).
- URL constants, token extraction, API GET/POST methods, and Aliyun request construction otherwise align with `internal/extractor/haiyangknow/SOURCE_ALIGN.md`.
- No response-body leak found.

## haozaixian

NO ISSUE.

Evidence:
- URL constants and single/AI course API flows match `internal/extractor/haozaixian/SOURCE_ALIGN.md`.
- Direct POST helper closes response bodies; reviewed error returns are propagated, and no nil panic or dead import was found.

## houda

Issues:
- Missing source URL constant: source defines `learn_course_page_url = .../api/center/myOnlineCourse/getLearnCoursePage` (`Houda_Course.py:38`), but Go constants omit it (`internal/extractor/houda/houda.go:31-36`).
- Declared-but-unexecuted source flows: Go declares `urlCourseList`, `urlStageLaw`, `urlLiveDetail`, and `urlMaterial` (`internal/extractor/houda/houda.go:31-35`), but call-site search only finds `urlLearnCourse` being posted (`internal/extractor/houda/houda.go:137`). Source uses course list selection (`Houda_Course.py:347-358`, `Houda_Course.py:397-400`) and stage-law detail (`Houda_Course.py:462-470`). As written, Go requires `classId` in the URL and errors if absent (`internal/extractor/houda/houda.go:77-80`), so source course-list/stage/material branches are not actually aligned.
- CSSLCloud resolution is correctly delegated to shared helper paths; no direct response-body leak found.

## houdu

NO ISSUE.

Evidence:
- API format, POST JSON method, lesson/play/material JSON key handling, and Baijiayun delegation align with `internal/extractor/houdu/SOURCE_ALIGN.md`.
- Direct POST helper closes response bodies; no nil panic, unchecked material error, unused import, or dead code found in reviewed paths.

## hqwx

NO ISSUE.

Evidence:
- URL constants, auth/cookie aliases, GET/POST loaders, route modes, resource/subtitle/plan JSON keys, and HLS key rewrite match `internal/extractor/hqwx/SOURCE_ALIGN.md`.
- `return nil, nil` sites reviewed are no-match helper returns, not unchecked errors or nil dereferences. Direct GET response body is closed.

## htknow

Issues:
- CRITICAL: source answer/quest endpoints are missing from Go. Source defines `answer_tag_url`, `answer_num_url`, `answer_list_url`, and `answer_create_paper_url` (`Htknow_Course.py:45-48`), but Go constants stop at video/detail endpoints (`internal/extractor/htknow/htknow.go:20-31`). The answer branch is therefore absent despite `SOURCE_ALIGN.md` saying no blockers.
- CRITICAL: HTML-only course items are dropped. Source preserves `html_content`, `product_type`, `product_token`, and `video_name` for single/column/series items (`Htknow_Course.py:291-307`, `Htknow_Course.py:373-381`, `Htknow_Course.py:418-427`). Go carries `html` into `source`, but `mediaFromSources` only appends entries when `src.url != ""` (`internal/extractor/htknow/htknow.go:343-358`). Pure 图文 / pay-content entries can be lost and reported as `no playable video URL`.
- POST JSON helper closes response bodies; no direct leak found.

## huatu

NO ISSUE.

Evidence:
- URL constants, Huatu headers/token aliases, GET query encoding, syllabus/player/VOD/Baijiayun flows, JSON keys, material URL quoting, and top-level `Entries` behavior match `internal/extractor/huatu/SOURCE_ALIGN.md`.
- Helper `return nil, nil` sites reviewed are no-match paths; direct request helpers propagate errors and close bodies.
