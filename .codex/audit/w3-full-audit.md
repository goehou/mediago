# w3 full audit of ledu batch

Scope: `huke88 icourse163 icourses icve imooc itbaizhan jianshe99 jinbangshidai jingtongxue kaimingzhixue kaoyanvip keqq koolearn kuke ledu`.

Method: compared each Go extractor with the decompiled Python source under `~/code/xwz-downloader-source-release/decompiled_full/Mooc/Courses/`, then scanned the target Go files for nil-panic patterns, response-body leaks, unchecked errors, dead imports, and stub/unsupported behavior.

Summary:

| site | result |
|---|---|
| huke88 | NO ISSUE |
| icourse163 | CRITICAL issues |
| icourses | NO ISSUE |
| icve | NO ISSUE |
| imooc | CRITICAL issues |
| itbaizhan | NO ISSUE |
| jianshe99 | NO ISSUE |
| jinbangshidai | NO ISSUE |
| jingtongxue | NO ISSUE |
| kaimingzhixue | Issues |
| kaoyanvip | NO ISSUE |
| keqq | NO ISSUE |
| koolearn | Issues |
| kuke | NO ISSUE |
| ledu | Minor code-review issue |

## huke88

NO ISSUE

- Source alignment: `internal/extractor/huke88/SOURCE_ALIGN.md` matches the Python constants and flows from `Huke88_Base.py` / `Huke88_Course.py`: `login_check_url`, `course_url`, `study_url`, `purchased_study_url`, `video_play_url`, `file_url`, cookie `_identity-usernew`, HTML/API headers, course page GET, purchased-study GET pagination, video/file POST forms, and confirm retries.
- Code review: no nil panic, direct response-body leak, dead import, fabricated URL, or material unchecked error found in the huke88 target files.

## icourse163

### CRITICAL: source-supported Icourse163 URL families are rejected or absent

- Source evidence: `Mooc_Config.pyc.1shot.cdc.py:301-305` registers `Icourse163_App`, `Icourse163_Textbook`, `Icourse163_Mooc`, `Icourse163_Column`, and `Icourse163_Kaoyan`.
- Source evidence: `Icourse163_Kaoyan.pyc.1shot.cdc.py:31-35` defines the Kaoyan course, term, live, and pay APIs.
- Source evidence: `Icourse163_Column.pyc.1shot.cdc.py:29-32` defines column APIs; `Icourse163_Textbook.pyc.1shot.cdc.py:40-44` defines textbook APIs.
- Go evidence: `internal/extractor/icourse163/icourse163.go:10-12` documents that only the common `/course/CID-NNN[?tid=MMM]` flow is implemented and other subflows are rejected.
- Go evidence: `internal/extractor/icourse163/icourse163.go:67-77` explicitly returns an unsupported-flow error for `/learn/kaopei-...`; no column/textbook/app/package handlers are registered in this site package.
- Impact: URLs accepted by the Python downloader are hard-fail paths in the Go extractor. This is stub-like source coverage loss, so it is marked `CRITICAL`.

### Issue: MOOC course-list, join/access, subtitle, and attachment flows are incomplete

- Source evidence: `Icourse163_Mooc.pyc.1shot.cdc.py:39-48` defines `join_url`, `course_list_url`, `new_infos_url`, `sub_url`, `attach_url`, and `timestamp_url`.
- Source evidence: `Icourse163_Mooc.pyc.1shot.cdc.py:288-326` pages learned courses; `:357-388` checks term access and joins a course; `:742-760` POSTs the subtitle API; `:984-988` builds attachment downloads.
- Go evidence: `internal/extractor/icourse163/icourse163.go:29-39` declares only the common MOOC constants; `subURL` is declared but the code path does not call it. There is no `join_url`, `course_list_url`, `new_infos_url`, or `attach_url` equivalent.
- Impact: source-visible learned-course selection, enrollment/access recovery, subtitle fallback, and non-video resources can be missed.

### Code review issue: selected network errors are swallowed

- Go evidence: `internal/extractor/icourse163/icourse163.go:111` ignores the `home.htm` fetch error when deriving `memberID`; `:266` ignores `getLessonUnitLearnVo` POST errors; `:284` ignores timestamp fetch errors.
- Impact: these do not create an immediate nil panic, but they can convert real request failures into silent empty streams or fallback behavior, making diagnosis and source-equivalence weaker.

## icourses

NO ISSUE

- Source alignment: `internal/extractor/icourses/SOURCE_ALIGN.md` matches `Icourse_Base`, `Icourse_Cuoc`, and `Icourse_Mooc`: API root, Bearer token/cookie aliases, GET JSON helper, CUOC detail/resource APIs, MOOC detail/chapter/resource/sub/doc APIs, and resource normalization for video/document/attachment types.
- Code review: no nil panic, direct response-body leak, dead import, fabricated URL, or material unchecked error found. `internal/extractor/icourses/resources.go:105` returns `nil, nil` for a category with no resources and is not a panic or stub path.

## icve

NO ISSUE

- Source alignment: `internal/extractor/icve/SOURCE_ALIGN.md` matches the AI ICVE endpoints, `courseInfoId/courseId/parentId` parameters, status URL, JSON keys (`courseName`, `schoolName`, `ossOriUrl`, `ossGenUrl`, `fileType`, `fileUrl`), and source quality selection.
- Code review: no nil panic, dead import, fabricated URL, or material unchecked error found. The direct `c.Get` path in `internal/extractor/icve/icve.go` closes and drains `resp.Body`.

## imooc

### CRITICAL: core `imooc_decode` playback transform is not implemented in Go

- Source evidence: `Imooc_Config.pyc.1shot.cdc.py:23-34` loads a JS function and calls `imooc_decode`.
- Source evidence: `Imooc_Class.pyc.1shot.cdc.py:404-431`, `Imooc_Code.pyc.1shot.cdc.py:240-267`, and `Imooc_Free.pyc.1shot.cdc.py:170-191` all fetch an encoded `info` value, decode it, follow the selected video URL, and decode the second `info` value.
- Go evidence: `internal/extractor/imooc/imooc.go:14-17` states the JS decode is unavailable, and `:90` returns an unsupported paid-content error instead of decoding.
- Impact: paid class/coding/free lesson playback that the Python source resolves cannot be resolved by Go. This is stub-like core playback behavior, so it is marked `CRITICAL`.

### Issue: class/coding lifecycle requests do not match the source bodies and side calls

- Source evidence: `Imooc_Class.pyc.1shot.cdc.py:34-39` and `Imooc_Code.pyc.1shot.cdc.py:32-37` define `m3u8_url`, `start_media_url`, `start_learn_url`, `end_learn_url`, and `end_media_url`.
- Go evidence: `internal/extractor/imooc/imooc.go:60-66` POSTs only `mid`, `cid`, and `_id`, never calls `start_media_url` / `end_media_url`, and ignores both lifecycle errors.
- Impact: source session setup/teardown semantics, including media-info and record/duration-dependent bodies, are not faithfully reproduced.

### Issue: course-level tree parsing is missing

- Source evidence: `Imooc_Class.pyc.1shot.cdc.py:140-176` parses `class.imooc.com/sc/{cid}/learn` HTML into the course tree; `Imooc_Free.pyc.1shot.cdc.py:100-143` parses free-course chapters, videos, codes, and exercises.
- Go evidence: `internal/extractor/imooc/imooc.go:51-54` only extracts `cid` and optional `mid` from the URL; `mediaURL()` can be called with an empty `mid` for course-level URLs.
- Impact: source-supported course pages with multiple lessons are not enumerated and can fail before reaching any real lesson.

## itbaizhan

NO ISSUE

- Source alignment: `internal/extractor/itbaizhan/SOURCE_ALIGN.md` matches `navlist_url`, `stage_url`, `play_url`, `check_url`, `course_list_url`, referer/origin/UA, GET/JSON methods, stage/rightlist fields, HTML play-info extraction, and Polyv secure resolution through shared helpers.
- Code review: no nil panic, response-body leak, dead import, fabricated URL, or material unchecked error found. Tuple returns like `return nil, nil, err` are valid multi-return paths, not nil dereferences.

## jianshe99

NO ISSUE

- Source alignment: `internal/extractor/jianshe99/SOURCE_ALIGN.md` matches the Jianshe99 constants, video-list/replay payload APIs, replay JSON tags, and CSSLcloud login/play-info/m3u8 rewrite through `shared.CssLcloudResolvePlayInfo` and `shared.CssLcloudRewriteM3U8Keys`.
- Code review: no nil panic, response-body leak, dead import, fabricated URL, or material unchecked error found.

## jinbangshidai

NO ISSUE

- Source alignment: `internal/extractor/jinbangshidai/SOURCE_ALIGN.md` matches `api_base`, student info, course list/detail/play/token endpoints, cookie token/JWT device handling, POST JSON flow, syllabus traversal, and Baijiayun helper usage.
- Code review: no nil panic, response-body leak, dead import, fabricated URL, or material unchecked error found. Direct POST response bodies are closed in `postJSON`.

## jingtongxue

NO ISSUE

- Source alignment: `internal/extractor/jingtongxue/SOURCE_ALIGN.md` matches base domain/header behavior, relative API paths, course list/status pagination, detail/chapter/lecture/play-param requests, JSON keys, direct playback URL extraction, and BokeCC resolution through shared helper.
- Code review: no nil panic, response-body leak, dead import, fabricated URL, or material unchecked error found.

## kaimingzhixue

### Issue: course file/material nodes from the Python source are not emitted

- Source evidence: `Kaimingzhixue_Course.pyc.1shot.cdc.py:417-439` defines `_parse_file_info`; `:533-539` appends `datum` / `files` entries through that parser.
- Go evidence: `internal/extractor/kaimingzhixue/kaimingzhixue.go:255-295` only appends `Kind: "video"` and `Kind: "live_playback"`; `kzxItem` has no file URL/name/format fields.
- Impact: courseware/material attachments present in the source downloader are silently dropped by Go.

### Issue: public `courseBasis` price/order-price path is declared but unused

- Source evidence: `Kaimingzhixue_Course.pyc.1shot.cdc.py:33` defines `public_course_url = 'https://www.lckmzx.com/api/app/courseBasis'`; `:226-279` uses it to page public-course metadata and set price/purchase/title data.
- Go evidence: `internal/extractor/kaimingzhixue/kaimingzhixue.go:34` declares `urlPublicCourse`, but no call site uses it; `:197-218` only loads `myStudy/{course_type}` lists.
- Impact: public-course metadata enrichment is missing, and the current `SOURCE_ALIGN.md` overstates HTTP-flow coverage for this constant.

## kaoyanvip

NO ISSUE

- Source alignment: `internal/extractor/kaoyanvip/SOURCE_ALIGN.md` matches user-info/cookie flow, mycourse/delivery/uuid outline APIs, VOD hls.videocc.net URL construction, key-token rewrite, MP4 fallback, live-record regex, timestamp/sign flow, and Polyv live URL parsing.
- Code review: no nil panic, response-body leak, dead import, fabricated URL, or material unchecked error found.

## keqq

NO ISSUE

- Source alignment: `internal/extractor/keqq/SOURCE_ALIGN.md` matches `get_plan_list`, course page `__NEXT_DATA__`, `get_terms_detail`, chapter `task_info` parsing, `describe_rec_video`, subtitles, and the DRM token construction.
- Code review: no nil panic, response-body leak, dead import, fabricated URL, or material unchecked error found. Optional fallback errors are contained by subsequent explicit errors when no media can be produced.

## koolearn

### Issue: registered extractor narrows Koolearn to direct Roombox class URLs and skips source discovery flows

- Source evidence: `Koolearn_Base.pyc.1shot.cdc.py:31-33` defines order index/detail pages and `:183-196` pages orders for price/purchase state.
- Source evidence: `Koolearn_App.pyc.1shot.cdc.py:34-35` defines `study.koolearn.com` and `my-data?type=...`; `:92-148` uses that app list to select a course.
- Source evidence: `Koolearn_Course.pyc.1shot.cdc.py:32-41` defines study-course, category, lesson, VOD, and m3u8 APIs; `:184-260` parses course stages/categories/lessons.
- Source evidence: `Koolearn_Roombox.pyc.1shot.cdc.py:37-41` defines Roombox class/schedule/lesson/playback APIs and has `_get_course_list` / `_get_course_list_again` before direct lesson resolution.
- Go evidence: `internal/extractor/koolearn/koolearn.go:61-99` requires `XDF_H5_TOKEN`, parses a `classId` directly from the input URL, calls only `fetchRoomboxLessons`, and errors if no class id is present.
- Go evidence: `internal/extractor/koolearn/koolearn.go:31-39` declares `urlOrderIndex`, `urlOrderDetail`, `urlStudyHome`, `urlMyData`, `urlRoomCourse`, and `urlRoomSchedule`, but targeted search finds no call sites beyond declarations/comment.
- Impact: Koolearn URLs/account-home invocations that Python resolves through order/my-data/study/Roombox course-list selection fail in Go unless the caller already supplies a direct Roombox classroom id.

### Code review issue: playback fallback error is swallowed

- Go evidence: `internal/extractor/koolearn/koolearn.go:167-171` ignores `fetchRoomboxModuleURL` errors and drops the lesson if no URL is found.
- Impact: this avoids a nil panic, but it can hide per-lesson API failures and silently omit lessons from the returned playlist.

## kuke

NO ISSUE

- Source alignment: `internal/extractor/kuke/SOURCE_ALIGN.md` matches Kuke base/course URL constants, signed POST flow, cookie validation, order/course/SVIP/detail/subcourse APIs, Polyv node-info request, and Polyv m3u8/key handling.
- Code review: no nil panic, response-body leak, dead import, fabricated URL, or material unchecked error found. Direct signed POST responses are closed.

## ledu

### Minor code-review issue: cookie-validation request error is ignored

- Source evidence: `Ledu_Base.pyc.1shot.cdc.py:2004-2022` treats the PC cookie validation request as a success/failure gate.
- Go evidence: `internal/extractor/ledu/ledu.go:79-80` issues the corresponding `courseSubjectList` request but discards both result and error.
- Impact: this does not create a nil panic and later class/detail requests still produce explicit failures, but the source validation gate is weaker and a bad cookie can fail later with less precise diagnostics.

Source alignment otherwise matches `internal/extractor/ledu/SOURCE_ALIGN.md`: class list, course detail list, video info, handout PDF, flexible payload extraction, and media/file URL key handling are present.
