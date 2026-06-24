# worker-6 full extractor audit

Scope: `xiwang xsteach xueersi xuelang xuetang yangcong yixiaoerguo yizhiknow youdao youyuan youzan zhaozhao zhengbao zhihuishu zlketang`.

Source root: `~/code/xwz-downloader-source-release/decompiled_full/Mooc/Courses/`.

Review checks applied per site:
- URL constants: byte-level comparison against `.cdc.py` literals, allowing `{cid}` -> `%s` / format placeholder conversion.
- HTTP flow: GET/POST choice and source request order.
- JSON keys: Go struct tags / map lookups against source `dict.get(...)` keys.
- Auth: cookie names, referer/origin/header families.
- Code review: nil panic, unclosed response body, unchecked errors, dead code / unused import.

Summary:
- CRITICAL: `zhihuishu` is still an explicit direct-video-only partial/stub path for normal course URLs.
- No unclosed response body was found in the audited direct `c.Post` sites; `util.GetString`, `util.GetBytes`, and `util.PostForm` close bodies.
- `go build ./...` is the unused-import/dead-import proof; individual low-severity blank-error findings are listed below.

## xiwang

Issues:
- MEDIUM: same source directory contains the `xiwang.youke` / `xiwang.suyang` variants, but the Go extractor is hard-wired to the `xiwang.com` endpoint family and pattern only. Source evidence: `Xiwang_Youke.pyc.1shot.cdc.py:40-47` uses `i.bcc.wen-su.com`, `studentlive.bcc.wen-su.com`, `api.xue.wen-su.com`; `Xiwang_Base.pyc.1shot.cdc.py:261-266` switches login check between `api.xue.xiwang.com`, `api.xue.xi-xue.com`, and `api.xue.wen-su.com`. Go evidence: `internal/extractor/xiwang/xiwang.go:17-26` only defines the `xiwang.com` family and `xiwang.go:30` only matches `xiwang.com|bcc.xiwang.com`.

No CRITICAL issue: the declared `xiwang.com` URL constants, POST/GET methods, JSON keys, and cookie/referer headers match the main `Xiwang_Course` / `Xiwang_Base` source path.

## xsteach

Issues:
- LOW: unchecked error return is deliberately dropped at `internal/extractor/xsteach/xsteach.go:73` (`courses, _ := fetchCourses(c, h)`). The current `fetchCourses` implementation never returns a non-nil error, but this still leaves a dead/unused error contract and violates the audit's unchecked-error rule.

No CRITICAL issue: source URLs `my-course/list-v3`, `my-course-combobox`, `course-detail`, `period`, `get-period-list`, `vod/period/play`, `teach-coach/play`, `live/enter/play`, and QCloud `getplayinfo` match `internal/extractor/xsteach/xsteach.go:22-33`; the flow is GET+JSON as in `Xsteach_Course.pyc.1shot.cdc.py:45-53` and request methods around `:167`, `:252-264`, `:323`, `:440`, `:460`.

## xueersi

Issues:
- LOW: `postJSON` ignores `json.Marshal` and `io.ReadAll` errors at `internal/extractor/xueersi/xueersi.go:233` and `:239`. The payloads are simple maps, so this is unlikely to trigger, but it is still an unchecked error return.

No CRITICAL issue: URL constants and methods match the source (`Xueersi_Base.pyc.1shot.cdc.py:30-31`, `Xueersi_Course.pyc.1shot.cdc.py:33-37`, plus `.das` vodshow/drama constants); all direct response bodies in `postJSON` are closed at `xueersi.go:238`.

## xuelang

Issues:
- HIGH: DRM m3u8 key handling is incomplete. Source decrypts the key and rewrites the m3u8 text before downloading (`Xuelang_Course.pyc.1shot.cdc.py:271-291`, `:297-309`). Go fetches the same token/key endpoints but discards the decrypted key at `internal/extractor/xuelang/xuelang.go:231-233`, and `playMedia` / `media()` only return the original m3u8 URL plus metadata (`xuelang.go:44-47`, `helpers.go:59-60`). Encrypted variants can therefore fail downstream even though URL constants and GET methods align.
- LOW: helper `postJSON` ignores `json.Marshal` and `io.ReadAll` errors at `internal/extractor/xuelang/helpers.go:21` and `:27`.

No CRITICAL issue: no stub, fabricated URL, nil panic, or unclosed response body found.

## xuetang

Issues:
- HIGH: URL/source coverage is partial for the site directory. Go only implements `/api/v1/lms/learn/product/info`, `/learn/course/chapter`, `/learn/leaf_info`, and `/service/playurl` (`internal/extractor/xuetang/xuetang.go:68-176`). Source also defines and uses course fallback / price / join / live / training endpoints: `url_couse_info`, `url_liveVideo`, `url_lang`, `url_join`, `url_product`, `url_train` in `Xuetang_Course.pyc.1shot.cdc.py:34-43`; live single extraction `/api/v1/lms/learn/live_info/{}/{}/?sign={}` in `Xuetang_Live.pyc.1shot.cdc.py:32`, `:64-91`; training classroom lookup `/api/v1/lms/learn/training/camp/classrooms/?sign={}` in `Xuetang_Train.pyc.1shot.cdc.py:32`, `:45-64`. Current `patterns` only cover `course|learn` paths, not `live` / `training` (`xuetang.go:24-34`).
- MEDIUM: source falls back from `learn/course/chapter` to `/api/v1/lms/product/get_course_detail/?cid={cid}` when the signed chapter request fails (`Xuetang_Course.pyc.1shot.cdc.py:230-244`); Go returns immediately on `course/chapter` error at `xuetang.go:74-77` and does not implement this fallback.
- LOW: title request error is ignored at `xuetang.go:68`; a failed `product/info` request silently degrades to `xuetang_<cid>`.

No CRITICAL issue: no explicit stub text, fabricated URL, nil panic, or body leak found.

## yangcong

Issues:
- LOW: source warm-up/order-auth calls are made but their errors are discarded at `internal/extractor/yangcong/yangcong.go:68-69`. This matches non-decisive warm-up behavior but still violates the unchecked-error review item.
- LOW: POST helper ignores `json.Marshal` error at `yangcong.go:216`. Response body is closed and `io.ReadAll` is checked at `yangcong.go:225-226`.

No CRITICAL issue: `school-api.yangcong345.com` endpoints, GET/POST split, JSON keys, and auth header/cookie handling match `Yangcong_Base` / `Yangcong_Course` source and `SOURCE_ALIGN.md`.

## yixiaoerguo

Issues:
- LOW: POST helper ignores `json.Marshal` error at `internal/extractor/yixiaoerguo/yixiaoerguo.go:227`; response body is closed and read errors are checked at `:232-236`.
- LOW: audition-unlock fallback error is intentionally ignored at `yixiaoerguo.go:304`, so a failed unlock is indistinguishable from “not needed” during later resolution.
- LOW: `qxReplaySVRURL` and `qxHLSEncryptURL` are kept only via blank assignments at `yixiaoerguo.go:371-372`; this is dead runtime code unless the qianxue replay/HLS-encrypt path is later implemented.

No CRITICAL issue: signed API headers, Biguo endpoints, qianxuecloud URL constants, GET/POST split, and JSON key paths match the source alignment document.

## yizhiknow

Issues:
- LOW: curriculum status warm-up result and error are ignored at `internal/extractor/yizhiknow/yizhiknow.go:88`.
- LOW: signed POST helper ignores `json.Marshal` error at `yizhiknow.go:203`; response body is closed and read errors are checked at `:208-212`.

No CRITICAL issue: API host, signing secret, token aliases, paths, GET/POST split, and JSON keys match `Yizhiknow_Base` / `Yizhiknow_Course` and `SOURCE_ALIGN.md`.

## youdao

NO ISSUE.

Evidence: URLs `user_status.jsonp`, two `my-orders`, two `products/after-sale`, two `products/after-sale/{cid}`, and `hikari-live/api/consumer/v1/key` match source (`Youdao_Base.pyc.1shot.cdc.py:32-34`, `Youdao_Shengxue.pyc.1shot.cdc.py:37-41`) and Go constants (`internal/extractor/youdao/youdao.go:18-26`). The extractor uses source-aligned GET+JSON/key-fetch flow, closes bodies through `util`, and I found no nil panic, body leak, unused import, or actionable unchecked network error.

## youyuan

NO ISSUE.

Evidence: URLs `getByCourseId`, `listPresentOrPrevious`, `getToken`, and Baijiayun `getPlayUrl` match source (`Youyuan_Course.pyc.1shot.cdc.py:30-33`) and Go constants (`internal/extractor/youyuan/youyuan.go:18-22`). GET flow and `data.courseName`, `courseLessonList`, `videoId`, `token` JSON keys match `SOURCE_ALIGN.md`; HTTP bodies are closed by `util` / shared Baijiayun helper.

## youzan

NO ISSUE.

Evidence: source relative APIs `/wscvis/course/detail/goods.json`, `/wscvis/knowledge/getColumnChapters.json`, `/wscvis/knowledge/contentAndLive.json`, `/wscvis/course/getSimple.json`, live-link APIs, room API, and asset-state API match `internal/extractor/youzan/youzan.go` and `SOURCE_ALIGN.md`. Flow is GET+JSON as source `_request_json`, with matching alias/kdtId/referer handling. No nil panic, response leak, unused import, or actionable unchecked error was found.

## zhaozhao

Issues:
- MEDIUM: `myBuyProductList` is both source auth/course-list probe and first payload source, but its error is discarded at `internal/extractor/zhaozhao/zhaozhao.go:223`. This can hide an auth/signature failure and continue with less context.
- LOW: play-safe token payload marshal error is ignored at `zhaozhao.go:416`. Payload is a simple string map, so runtime risk is low.

No CRITICAL issue: yikao88 signed headers, product/package/detail URLs, Polyv secure/key constants, and play-token fallback URLs match the source alignment document; no body leak or nil panic found.

## zhengbao

Issues:
- LOW: `postJSON` ignores `json.Marshal` and `io.ReadAll` errors at `internal/extractor/zhengbao/zhengbao.go:426` and `:438`. Response body is closed at `:437`.

No CRITICAL issue: doorman URLs/crypto flow, `cdeluid`/`sid` auth extraction, elearning material/video URLs, GET/POST split, and JSON/HTML regex keys match `Zhengbao_*` source and `SOURCE_ALIGN.md`; no nil panic or unused import found.

## zhihuishu

Issues:
- CRITICAL: extractor is still an explicit direct-video-only partial implementation for normal course URLs. Go returns `zhihuishu course-tree traversal needs HTML scraping that's not implemented` when no `videoID` is present (`internal/extractor/zhihuishu/zhihuishu.go:46-49`). The same file comments admit course traversal is blocked (`zhihuishu.go:9-11`). This violates the no-stub rule even though `verify_full_alignment.py` classifies it as PASS by shape.
- CRITICAL: source URL coverage is far from byte-aligned. Go only calls `https://newbase.zhihuishu.com/video/initVideo?videoID=%s`, `https://newbase.zhihuishu.com/video/changeVideoLine?videoID=%s&lineID=%d&uuid=%s`, and uses `https://onlineweb.zhihuishu.com/` as referer (`zhihuishu.go:53`, `:77-109`). Source directory contains course/school/interest/smart/live flows and endpoints including `initVideoNew`, `initVideoToC`, `ai-course-platform...queryCourseResourceInfo`, `coursehome...queryPreviewFilePath`, `b2cpush...query2CCourseInfo`, `b2cpush...query2CCourseCatalog`, `hiexam-server...findCourseInfo`, `studyresources...queryResourceTree`, and live VOD room URLs (`Zhihuishu_Smart.pyc.1shot.cdc.py:77-89`, `Zhihuishu_Interest.pyc.1shot.cdc.py:34-38`, `Zhihuishu_School.pyc.1shot.cdc.py:32-37`, `Zhihuishu_Live.pyc.1shot.cdc.py:112-144`).

No nil panic or body leak was found in the implemented direct-video path, but the extractor is not acceptable as a source-aligned full site implementation.

## zlketang

Issues:
- MEDIUM: auth probe result is ignored at `internal/extractor/zlketang/zlketang.go:118` (`ctx.requestJSON(checkURL, nil, refererURL)`). Source `_check_cookie` treats `user_info` success (`errcode == 0` or `code == 200` with `data`) as the authentication decision (`Zlketang_Base.pyc.1shot.cdc.py:222-228`). Current Go may proceed with invalid cookies and only fail later.

No CRITICAL issue: wxpub API constants, web parameter keys, live/QCloud flow, AES/RSA helpers, JSON key families, referer headers, and body-close behavior match `Zlketang_*` source and `SOURCE_ALIGN.md`; no nil panic or unused import found.
