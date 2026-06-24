# w1 full audit

Scope: audited `internal/extractor/<site>` against `~/code/xwz-downloader-source-release/decompiled_full/Mooc/Courses/<Site>/` for these 16 sites: `ahu`, `aishangke`, `baijiayunxiao`, `bilibili`, `caixuetang`, `cctalk`, `cctv`, `chaoge`, `chaoxing`, `ckjr`, `classin`, `cnmooc`, `cto51`, `dingtalk`, `dongao`, `douyin`.

Audit boundary: source alignment was checked against the extractor-visible HTTP/API flow and JSON/auth handling. Local/offline render subsystems in the Python tree were not treated as failures unless the Go extractor claimed or routed into that flow. `CRITICAL` is used for stub/blocked implementation, fabricated source URL, or nil-panic class risk.

Verification run:

- `python3 scripts/verify_full_alignment.py`: target sites all PASS in the script output; global summary `PASS 91`, `BLOCKED 1 (yikaobang)`, `STUB 0`, `NO_EXTRACT 0`.
- `go test ./internal/extractor/{ahu,aishangke,baijiayunxiao,bilibili,caixuetang,cctalk,cctv,chaoge,chaoxing,ckjr,classin,cnmooc,cto51,dingtalk,dongao,douyin}`: PASS.
- `go vet ./internal/extractor/{ahu,aishangke,baijiayunxiao,bilibili,caixuetang,cctalk,cctv,chaoge,chaoxing,ckjr,classin,cnmooc,cto51,dingtalk,dongao,douyin}`: PASS.
- `xwz-dl doctor`: `ok=true`.

## ahu

NO ISSUE

- Source URLs `course_list_url`, `course_info_url`, `video_play_url` match `Ahu_Course.pyc.1shot.cdc.py:38-40` with `{cid}/{lesson_id}` converted to `%s` in `internal/extractor/ahu/ahu.go:23-25`.
- HTTP method is GET throughout (`GetString`); auth/referer flow is copied through `Referer`/`referer` headers and cookie jar.
- Aliyun JSON tags and play-info fields match the source key chain documented in `internal/extractor/ahu/SOURCE_ALIGN.md`.
- Code review: no raw response leak, no unchecked request construction, no dead imports; the ignored `hmac.Hash.Write` return at `ahu.go:265` is safe because hash writes cannot fail.

## aishangke

NO ISSUE

- Source URLs for course detail, series list, enter course, and CSSLCloud replay endpoints match `Aishangke_Course.pyc.1shot.cdc.py` and `internal/extractor/aishangke/aishangke.go`.
- CSSLCloud resolution correctly uses `shared.CssLcloudResolvePlayInfo` and does not inline the platform login/play API.
- HTTP methods and auth headers match: GET for Loveshangke pages/APIs, source-aligned replay POST/GET via the shared helper, cookie jar plus `Referer`/`Origin`.
- Code review: no raw response leak, no nil-panic error-use pattern, no dead imports.

## baijiayunxiao

NO ISSUE

- Source URLs for `Baijiayun_Video` and `Baijiayunxiao_Course` are represented by `urlGetPlayInfo`, `urlGetPlayURL`, `urlCourseInfo`, `urlToken`, and `urlPlayToken`.
- HTTP method alignment holds: GET for play/course/token APIs, POST JSON for live-enter.
- JSON tags cover the source keys for `data.periods`, `data.chapter`, `video_id`, `room_id`, `token`, `classid`, and shared Baijiayun playback fields.
- Code review: the raw POST response in `resolveLiveEnter` is closed with `defer resp.Body.Close()` before reading; optional fallback `json.Unmarshal` at `baijiayunxiao.go:181` is non-fatal and backed by regex fallback.

## bilibili

- CRITICAL: fabricated source URLs in the Go extractor. The decompiled course/Gongfang sources use PUGV and mall/gongfang endpoints such as `https://api.bilibili.com/pugv/...`, `https://mall.bilibili.com/mall-c/order/detail?orderId={cid}`, and `https://gf.bilibili.com/order/hyg-download` (`Bilibili_Course.pyc.1shot.cdc.py:35-38`, `Bilibili_Gongfang.pyc.1shot.cdc.py:35-37,228`). `internal/extractor/bilibili/bilibili.go:78,133` instead resolves generic public-video endpoints `x/web-interface/view` and `x/player/playurl`, which are not in the provided `.cdc.py` sources.
- ISSUE: source coverage is incomplete. `Bilibili_Gongfang.pyc.1shot.cdc.py:110,119,138` parses `shipOrderDetails` and `shipOrderDetailsId` and calls `querydownloadurl`; the Go extractor has no mall/gongfang URL constants or JSON mapping for that flow.
- Code review: no unclosed raw response body found; `resolveShortURL` closes the response at `bilibili.go:263`. The unchecked regexp error at `bilibili.go:258` is benign because the pattern is a constant literal.

## caixuetang

NO ISSUE

- Source API host/referer/origin/appcode constants and form POST wrapper align with `Caixuetang_Base.pyc.1shot.cdc.py` and `internal/extractor/caixuetang/caixuetang.go`.
- Course list, playinfo, material play, and download task/info APIs are implemented with source-aligned POST form calls.
- JSON parsing is map/tree based and covers the source key families documented in `internal/extractor/caixuetang/SOURCE_ALIGN.md`.
- Code review: no raw response leak; `util.Client.PostForm` closes response bodies; no unchecked request-construction path or dead imports.

## cctalk

NO ISSUE

- Source constants `CCTALK_BASE_URL`, `CCTALK_CONTENT_API_V1/V11/V12`, `CCTALK_PCWEB_KEY`, `CCTALK_TENANT_ID`, and `CCTALK_USER_AGENT` match `Cctalk_Config.pyc.1shot.cdc.py:55-62` and `internal/extractor/cctalk/cctalk.go:15-21`.
- Course/group/series/video API methods align with `Cctalk_Course.pyc.1shot.cdc.py` and use the same GET/POST split through `requestAPI`.
- JSON parsing is dynamic map traversal and covers source keys for data/list extraction, node traversal, media URL, and title fallback.
- Code review: no raw response leak, nil-panic risk, or dead import found.

## cctv

- ISSUE: auth/header flow is not source-aligned. Source constructs `cookie`, `Accept`, `Origin`, `Referer`, and `User-Agent` in `Cctv_Course.pyc.1shot.cdc.py:53-58`, and both `_request_text` and `_request_json` pass `self.header` to `requests.get` (`:122-149`). Go only sends `Referer` for the page request and sends nil headers for the API request at `internal/extractor/cctv/cctv.go:32-50`.
- ISSUE: JSON candidate mapping is incomplete. Source bytecode names candidate branches `hls_url`, `video_url`, `chapters4`, `chapters3`, `chapters2`, and `chapters`; Go only decodes top-level `title`, `hls_url`, `video_url`, and `chapters_url` at `cctv.go:55-60`.
- Code review: no body leak or dead import found; all JSON decode errors are checked.

## chaoge

NO ISSUE

- Source course detail/file/series/room URLs match `Chaoge_Course.pyc.1shot.cdc.py` and `internal/extractor/chaoge/chaoge.go`.
- CSSLCloud replay flow correctly uses `shared.CssLcloudResolvePlayInfo` and `shared.CssLcloudRewriteM3U8Keys`, not an inline implementation.
- HTTP methods, cookies, `Referer`/`Origin`, and `ccInfo` JSON-like parsing match the source-aligned flow.
- Code review: no raw response leak, nil-panic error-use pattern, or dead import found.

## chaoxing

- ISSUE: source coverage is incomplete. Source defines direct video, live, and Yun file URLs: `url_source`, `url_live`, and `url_yun_file` in `Chaoxing_Course.pyc.1shot.cdc.py:47-51`; it also parses `attachments`, `liveId`, `statusUrl`, and `downloadUrl` in `Chaoxing_Course.pyc.1shot.cdc.py:1333-1353` and `Chaoxing_Mooc.pyc.1shot.cdc.py:330,908-928`. Go only resolves an `objectId` and fetches `https://mooc1.chaoxing.com/ananas/status/%s` at `internal/extractor/chaoxing/chaoxing.go:37-55`.
- ISSUE: JSON parsing only handles `filename`, `http`, and `hls` (`chaoxing.go:60-78`), but the source flow also handles live/material attachment keys.
- Code review: no raw response leak, nil-panic risk, or dead import found.

## ckjr

NO ISSUE

- Source `api_host`, `qcloud_play_api`, webversion/fromApp constants, and CKJR headers match `internal/extractor/ckjr/ckjr.go` and `helpers.go`.
- HTTP methods align: GET JSON for CKJR APIs and QCloud playinfo.
- Dynamic map parsing covers the source key variants for auth, qcloud IDs/signature, media URLs, and title fallback.
- Code review: no raw response leak, nil-panic risk, or dead import found.

## classin

NO ISSUE

- Source record/token/CDN constants match `internal/extractor/classin/classin.go`.
- HTTP method alignment holds: POST form for token, lesson record, activity record, and user record classes.
- JSON tags cover `error_info.errno` and `data.token`; recursive parsing covers `video`, `pm3u8`, `m3u8`, `Url/url`, and mp4 path variants.
- Code review: no raw response leak, nil-panic risk, or dead import found.

## cnmooc

NO ISSUE

- Source origin/referer/login/user-agent and item detail endpoint match `internal/extractor/cnmooc/cnmooc.go`.
- HTTP method alignment holds: GET for course/session pages and POST form for `/item/detail.mooc`.
- HTML/JSON key handling covers `courseId`, `courseOpenId`, `nodeId`, `itemId`, `node`, `mediaResources`, and media URL fallback keys.
- Code review: no raw response leak, nil-panic risk, or dead import found.

## cto51

NO ISSUE

- Source course, lesson, API, training/order, and qcloud URL constants match `internal/extractor/cto51/cto51.go`.
- HTTP method alignment holds for source GET/JSON page flows and QCloud playinfo resolution.
- Dynamic key handling covers lesson lists, train routes, Ali/QCloud auth, and media URL variants documented in `SOURCE_ALIGN.md`.
- Code review: no raw response leak, nil-panic risk, or dead import found.

## dingtalk

- CRITICAL: live replay path is a blocked stub. Source imports and calls `probe_live_replay`, `probe_public_live_share`, `probe_preview_dentry`, `probe_notable_record`, and `probe_ai_transcribe` (`Dingtalk_Video.pyc.1shot.cdc.py:40,435,497,542,582,619`; `Dingtalk_Live_Client.pyc.1shot.cdc.py:2919-3127`). Go parses the live IDs but returns a hard failure at `internal/extractor/dingtalk/dingtalk.go:64-65` instead of performing the LWP/probe flow.
- ISSUE: URL/API coverage is incomplete for documents too. Source includes `https://alidocs.dingtalk.com/api/doc/info`, `https://alidocs.dingtalk.com/api/document/data`, and `https://alidocs.dingtalk.com/nt/api/docs/preset/binary` (`Dingtalk_Live_Client.pyc.1shot.cdc.py:3610-3619`); Go only posts `https://alidocs.dingtalk.com/nt/api/docs/preset` in `dingtalk.go:35,76`.
- Code review: no raw response leak or dead import found, but the live replay extractor is not an implementation-complete path.

## dongao

NO ISSUE

- Source base/course URL constants match `internal/extractor/dongao/dongao.go`.
- HTTP method alignment holds: GET catalog/lecture pages, POST form for stage/detail/live APIs, and lecture fallback handling.
- Dynamic JSON parsing covers catalog chapter/lecture lists, detail lecture IDs, titles, and media URL candidates.
- Code review: no raw response leak, nil-panic risk, or dead import found.

## douyin

- CRITICAL: no decompiled source directory exists for Douyin under `~/code/xwz-downloader-source-release/decompiled_full/Mooc/Courses/`, so the requested source-alignment checks cannot be satisfied. The Go URLs `https://ttwid.bytedance.com/ttwid/union/register/`, `https://aweme.snssdk.com/aweme/v1`, and `https://www.iesdouyin.com/share/video/%s/` in `internal/extractor/douyin/douyin.go:23-27` are unsupported by the provided `.cdc.py` evidence.
- CRITICAL: nil-panic class risk from ignored `http.NewRequest` errors followed by immediate dereference: `douyin.go:92-94`, `115-116`, `138-140`, and `291-295`.
- ISSUE: unchecked body read errors at `douyin.go:101,129,150` can silently parse partial/failed responses.
- Code review: raw response bodies are closed on the observed paths (`defer resp.Body.Close()` or direct close), and no dead import was found.
