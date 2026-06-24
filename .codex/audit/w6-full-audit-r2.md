# worker-6 full extractor audit, round 2

Scope reused from round 1 report `./.codex/audit/w6-full-audit.md`:
`xiwang xsteach xueersi xuelang xuetang yangcong yixiaoerguo yizhiknow youdao youyuan youzan zhaozhao zhengbao zhihuishu zlketang`.

Round-2 focus:
1. Re-check whether round-1 CRITICAL findings were fixed by later worker changes.
2. Deep-scan nil panic, HTTP response body leaks, unchecked errors, and dead runtime code.
3. Record only current-state evidence from this worktree.

Current-state summary:
- The round-1 `zhihuishu` CRITICAL findings are **not fixed**. `internal/extractor/zhihuishu/zhihuishu.go:48` still returns an explicit `not implemented` error for normal course URLs without `videoID`, and current `git blame` shows no later worker changed that line after the earlier source-alignment rewrite.
- No HTTP response body leak was found in the audited packages. Direct `c.Post` call sites close `resp.Body`; shared `util.GetBytes`, `util.GetString`, and `util.PostForm` close bodies in `internal/util/http.go:50-64` and `:74-92`.
- `go vet ./internal/extractor/...` passes, so no compiler/vet-detected dead imports or obvious unreachable mistakes were found.
- New/deeper code-review risks found in round 2 are mostly LOW nil-panic or unchecked-error cases, plus the existing HIGH source-coverage problems in `xuelang` / `xuetang`.

## xiwang

R2 status: round-1 MEDIUM remains open.

Issues:
- MEDIUM: variant coverage still only implements main `xiwang.com`. Current Go constants remain `xiwang.com` only at `internal/extractor/xiwang/xiwang.go:17-26`, while source still has `wen-su.com` / `xi-xue.com` branch families in `Xiwang_Youke.pyc.1shot.cdc.py:40-47` and `Xiwang_Base.pyc.1shot.cdc.py:261-266`.

Deep scan:
- No nil panic found: `firstMatch` loops over `len(m)`, so nil regex result is safe (`xiwang.go:221-228`).
- No response body leak found; network reads go through `util.GetString` / `util.PostForm`.
- No new unchecked decisive error found beyond intentionally skipped fallback play API attempts inside `resolveLesson`.

## xsteach

R2 status: round-1 LOW remains open; one additional LOW nil-panic risk found.

Issues:
- LOW: stale/unchecked error contract remains at `internal/extractor/xsteach/xsteach.go:73` (`courses, _ := fetchCourses(c, h)`). `fetchCourses` still currently returns nil error only, but the ignored error remains misleading.
- LOW: `normalizeURL` can panic on malformed relative URLs from API payloads. It ignores `url.Parse(s)` at `internal/extractor/xsteach/helpers.go:169-171`; `net/url.Parse("/%zz")` returns nil URL with an error, and `base.ResolveReference(nil)` panics.

Deep scan:
- No response body leak found; all HTTP reads use `util.GetString`.
- No CRITICAL issue found in this package.

## xueersi

R2 status: round-1 LOW remains open; no new higher-severity issue found.

Issues:
- LOW: `postJSON` still ignores `json.Marshal` and `io.ReadAll` errors at `internal/extractor/xueersi/xueersi.go:233` and `:239`.
- LOW: malformed string `sectionResource` JSON is silently ignored at `xueersi.go:205-208`, which can hide a source payload parse failure and simply drop the recording URL.

Deep scan:
- No nil panic found in indexed slices: `chapters`, `chapters2`, `sections`, and `addrs` all have `len(...)` guards before `[0]` access (`xueersi.go:183-229`).
- No response body leak found; `postJSON` closes `resp.Body` at `xueersi.go:238`.

## xuelang

R2 status: round-1 HIGH remains open; no new higher-severity issue found.

Issues:
- HIGH: DRM m3u8 key handling is still incomplete. Go still discards `decryptM3U8Key` at `internal/extractor/xuelang/xuelang.go:231-233`; `playMedia` only carries `videoURL`, `audioURL`, `videoID`, `keyID`, and `size` (`xuelang.go:44-47`), so the decrypted key / rewritten m3u8 text is not exposed. Source still rewrites m3u8 key URI before download (`Xuelang_Course.pyc.1shot.cdc.py:271-309`).
- LOW: helper `postJSON` still ignores `json.Marshal` and `io.ReadAll` errors at `internal/extractor/xuelang/helpers.go:21` and `:27`.

Deep scan:
- No nil panic found in current indexed accesses: `external_video_infos[0]`, regex submatches, and split parts are length-guarded (`xuelang.go:192-196`, `:242-258`).
- No response body leak found; direct POST sites close `resp.Body` at `xuelang.go:163` and `helpers.go:26`.

## xuetang

R2 status: round-1 HIGH/MEDIUM/LOW remain open.

Issues:
- HIGH: source coverage is still partial. Current Go only implements product info, chapter, leaf info, and playurl paths (`internal/extractor/xuetang/xuetang.go:68-176`). It still does not cover source live/training/fallback/price/join endpoints from `Xuetang_Course.pyc.1shot.cdc.py:34-43`, `Xuetang_Live.pyc.1shot.cdc.py:32` and `:64-91`, or `Xuetang_Train.pyc.1shot.cdc.py:32` and `:45-64`.
- MEDIUM: `course/chapter` fallback to `/api/v1/lms/product/get_course_detail/?cid={cid}` remains missing. Source fallback is at `Xuetang_Course.pyc.1shot.cdc.py:230-244`; Go still returns immediately on `course/chapter` error at `xuetang.go:74-77`.
- LOW: title request error is still ignored at `xuetang.go:68`.

Deep scan:
- No nil panic found: regex submatch and stream slice accesses are guarded (`xuetang.go:191-212`).
- No response body leak found; HTTP calls use `util.GetString`.

## yangcong

R2 status: round-1 LOW findings remain open.

Issues:
- LOW: source warm-up / order-auth calls still discard result and error at `internal/extractor/yangcong/yangcong.go:68-69`.
- LOW: POST helper still ignores `json.Marshal` error at `yangcong.go:216`; `io.ReadAll` is checked and body is closed at `yangcong.go:225-228`.

Deep scan:
- No nil panic found. `parseCourseRequest` ignores `url.Parse(raw)` but checks `u != nil` before dereference (`yangcong.go:100-116`); fragment split is guarded by `strings.Contains`.
- No response body leak found.

## yixiaoerguo

R2 status: round-1 LOW findings remain open.

Issues:
- LOW: POST helper still ignores `json.Marshal` error at `internal/extractor/yixiaoerguo/yixiaoerguo.go:227`; response body is closed/read-checked at `:232-236`.
- LOW: audition unlock still discards error at `yixiaoerguo.go:304`, hiding whether unlock failed or was unnecessary.
- LOW: `qxReplaySVRURL` and `qxHLSEncryptURL` remain blank-assigned only at `yixiaoerguo.go:371-372`; they are source-trace constants but dead at runtime.

Deep scan:
- No nil panic found. `parseCID` guards empty submatches (`yixiaoerguo.go:115-122`); query parsing and qianxue JSON parsing check errors before use.
- No response body leak found.

## yizhiknow

R2 status: round-1 LOW findings remain open.

Issues:
- LOW: curriculum status warm-up still discards result and error at `internal/extractor/yizhiknow/yizhiknow.go:88`.
- LOW: signed POST helper still ignores `json.Marshal` error at `yizhiknow.go:203`; response body is closed/read-checked at `:208-212`.

Deep scan:
- No nil panic found. `parseCID` guards empty submatches (`yizhiknow.go:111-118`), and media URL `[0]` access is preceded by `len(urls) == 0` check (`yizhiknow.go:294-298`).
- No response body leak found.

## youdao

R2 status: round-1 NO ISSUE mostly holds; no CRITICAL/HIGH issue found.

Issues:
- LOW: `fetchKey` ignores `url.Parse(keyURL)` error at `internal/extractor/youdao/youdao.go:208`; this is a safe literal today, but it is still unchecked parse error handling.

Deep scan:
- No nil panic found. `parseCID` and key URI submatches are length-guarded (`youdao.go:98-108`, `:196-204`); the fixed two-element URL slice is created locally before `[0]/[1]` swap (`youdao.go:144-148`).
- No response body leak found; HTTP calls use `util.GetString` / `util.GetBytes`.

## youyuan

R2 status: round-1 NO ISSUE becomes LOW due dead trace assignment.

Issues:
- LOW: `bjyAPI` is kept only through `_ = fmt.Sprintf(bjyAPI, vid, token)` at `internal/extractor/youyuan/youyuan.go:151` while actual resolution is delegated to `shared.BaijiayunResolveVOD` at `:147`. This is harmless but dead runtime code.

Deep scan:
- No nil panic found. `parseCID` is length-guarded (`youyuan.go:83-88`); all API errors in the decisive chain are returned.
- No response body leak found; HTTP calls use `util.GetString` / shared helper.

## youzan

R2 status: round-1 NO ISSUE becomes LOW due defensive nil-panic risk.

Issues:
- LOW: malformed scheme-relative input can lead to nil URL panic after ignored parse errors. `configure` accepts a URL with `Host` but empty `Scheme` (`internal/extractor/youzan/youzan.go:82-90`), which can set `shopBase` to an invalid value like `://host`. Later `buildHeaders` ignores `url.Parse(raw)` and calls `jar.Cookies(u)` at `youzan.go:126-128`; `apiURL` also ignores `url.Parse(...)` and dereferences `u` at `youzan.go:148-159`. Normal `https://...` inputs are safe, but this is still a nil-panic edge.

Deep scan:
- No response body leak found; HTTP calls use `util.GetString`.
- No unchecked decisive network error found in the primary `goods.json` request; fallback media APIs are intentionally best-effort.

## zhaozhao

R2 status: round-1 MEDIUM/LOW remain open; one additional LOW nil-panic edge found.

Issues:
- MEDIUM: `myBuyProductList` remains both auth/course-list probe and first payload source, but its error is discarded at `internal/extractor/zhaozhao/zhaozhao.go:223`.
- LOW: play-safe token payload marshal error remains ignored at `zhaozhao.go:416`.
- LOW: `pickFormat` ignores `url.Parse(rawURL)` and dereferences `u.Path` at `zhaozhao.go:764-765`. File URLs are usually validated before this path, but malformed HTTP URLs from upstream payloads can make `url.Parse` return nil and panic.

Deep scan:
- No response body leak found; direct body-returning POST uses `util.PostForm` and shared Polyv helpers.
- No CRITICAL issue found.

## zhengbao

R2 status: round-1 LOW remains open.

Issues:
- LOW: `postJSON` still ignores `json.Marshal` and `io.ReadAll` errors at `internal/extractor/zhengbao/zhengbao.go:426` and `:438`. Response body is closed at `:437`.

Deep scan:
- No nil panic found in regex submatch usage: `openURLRe`, `videoIDRe`, `h5VarsRe`, and `attrRe` all check match length before indexing (`zhengbao.go:468-475`, `:566-579`, `:603-609`).
- No response body leak found.

## zhihuishu

R2 status: round-1 CRITICAL findings remain open, not resolved by later worker changes.

Issues:
- CRITICAL: normal course-tree extraction is still explicitly not implemented. Current code still returns `zhihuishu course-tree traversal needs HTML scraping that's not implemented (please pass a videoID URL)` when `extractVideoID` fails (`internal/extractor/zhihuishu/zhihuishu.go:46-49`). Fresh scan still finds the `not implemented` string only in this package.
- CRITICAL: source URL coverage remains direct-video-only. Current Go still only calls `newbase.zhihuishu.com/video/initVideo`, `newbase.zhihuishu.com/video/changeVideoLine`, and uses `onlineweb.zhihuishu.com` as referer (`zhihuishu.go:53`, `:77-109`). Source directory still includes course/school/interest/smart/live endpoint families such as `initVideoNew`, `initVideoToC`, `ai-course-platform...queryCourseResourceInfo`, `coursehome...queryPreviewFilePath`, `b2cpush...query2CCourseInfo`, `b2cpush...query2CCourseCatalog`, `hiexam-server...findCourseInfo`, and `studyresources...queryResourceTree` in the same evidence lines cited by round 1.

Deep scan:
- No nil panic found in the implemented direct-video path: regex submatches are length-guarded (`zhihuishu.go:129-136`) and line slice truncation checks `len(ids) > 2` before slicing (`zhihuishu.go:97-104`).
- No response body leak found; HTTP calls use `util.GetString`.

## zlketang

R2 status: round-1 MEDIUM remains open; one additional LOW unchecked parse issue noted.

Issues:
- MEDIUM: auth probe result remains ignored at `internal/extractor/zlketang/zlketang.go:118`. Source `_check_cookie` treats `user_info` success as the authentication decision (`Zlketang_Base.pyc.1shot.cdc.py:222-228`), but Go can continue with invalid cookies until a later failure.
- LOW: `decodeSignedData` silently treats non-numeric `sign` as `0` because it ignores `strconv.Atoi` error at `zlketang.go:491`; this can hide a signed-data format drift and return nil media.

Deep scan:
- No nil panic found in current regex/index paths: course/product regex submatches are length-guarded (`zlketang.go:173-179`); fragment split is guarded by `strings.Contains` (`zlketang.go:160-161`).
- No response body leak found; HTTP calls use `util.GetString`.
