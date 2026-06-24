# houda 源码对齐对照

## URL 常量

| .cdc.py 行 | houda.go 行/名 | 一致? |
|---|---|---|
| Houda_Base.py:31-32 `referer = 'http://www.houdask.com/'`, `origin = 'http://www.houdask.com'` | houda.go:32-33 `urlOrigin`, `urlHome` | ✓ |
| Houda_Course.py:35-41 父站课程/课节/直播详情/资料/CSSL 回放 API | houda.go:35-41 `urlCourseList` / `urlStageLaw` / `urlLearnCourse` / `urlLearnCoursePage` / `urlLiveDetail` / `urlMaterial` / `urlCCViewPlayback` | ✓, `{room_id}/{record_id}` -> `%s/%s` |
| Houda_Course.py:42-45 CSSLCloud replay URL | houda.go:42-45 `urlCsslLogin` / `urlCsslPlay` / `urlCsslMeta` / `urlCsslOrigin` | ✓ |
| Houda_Course.py:46-50 CSSLCloud device/tpl/terminal/material type | houda.go:46-50 `csslDeviceType` / `csslDeviceVersion` / `csslTpl` / `csslTerminal` / `materialServiceTyp` | ✓ |

## HTTP 调用

| 源码方法 (line) | Go 函数 (line) | method | 一致? |
|---|---|---|---|
| Houda_Base._check_cookie line 147, `ifLogin` 解析 `code == '1'` | houda.go:128 `checkHoudaCookie` | GET | ✓ |
| Houda_Course._request_houda lines 170-185 | houda.go:170 `requestHouda` | POST form | ✓, 兼容源码 raw form 并补 `data=json` fallback |
| Houda_Course._get_course_list lines 347-358, `getLearnFirstPage` | houda.go:212 `fetchHoudaCourseList`, 220 `parseHoudaCourseList`, 253 `chooseHoudaCourse` | POST form | ✓ |
| Houda_Course._get_stage_law_data lines 459-471, `getXxStageAndLawList` + `classId` | houda.go:279 `fetchHoudaStageLaw`, 721 `collectHoudaLawRefs` | POST form | ✓ |
| Houda_Course._get_lesson_list lines 477-487, `getLearnCourse` + `classId` | houda.go:146 `fetchHoudaLessons`, 610 `parseHoudaLessons` | POST form | ✓, 额外保留 `getLearnCoursePage` fallback |
| Houda_Course live detail const line 39 | houda.go:316 `hydrateHoudaLesson`, 334 `fetchHoudaLiveDetail` | POST/GET fallback | ✓, 仅在课节缺少回放字段时补详情 |
| Houda_Course._get_materials disasm lines 4787-5018, `getList` + `lawId/serviceType/classId` | houda.go:289 `fetchHoudaMaterials` | POST form | ✓ |
| Houda_Course._resolve_cc_callback_url / _parse_cc_info decrypted consts `room_id`, `recordId`, `viewertoken` | houda.go:483 `resolveHoudaCCCallback` | GET | ✓ |
| Houda_Course._get_cc_session / _get_cc_video_url decrypted chain | houda.go:455 `resolveHoudaCSSL` -> `shared.CssLcloudResolvePlayInfo` | POST+GET helper | ✓ |
| Houda_Course._download_media_url decrypted `.m3u8` branch | houda.go:523 `rewriteHoudaM3U8` -> `shared.CssLcloudRewriteM3U8Keys` | GET | ✓ |

## JSON 字段映射

| 源码 key 链 | Go struct tag | 一致? |
|---|---|---|
| result.get('data', {}).get('liveList', []) | `parseHoudaLessons` in houda.go:610-622 | ✓ |
| lesson.get('title'/'name'/'courseName') | `Title/Name/CourseName` `json:"title"/"name"/"courseName"` in houda.go:388-390 | ✓ |
| lesson.get('courseId'/'classId'/'type') | `CourseID/ClassID/Type` tags in houda.go:391-393 | ✓ |
| lesson.get('ccLiveId'/'roomId'/'mainRoomId'/'recordId') | `CCLiveID/RoomID/MainRoomID/RecordID` tags in houda.go:394-397 | ✓ |
| lesson.get('liveUrl'/'playbackUrl'/'playbackMp4'/'playbackMp3') | `LiveURL/PlaybackURL/PlaybackMP4/PlaybackMP3` tags in houda.go:398-401 | ✓ |
| lesson.get('stageId'/'stageName'/'lawId'/'lawName') | `StageID/StageName/LawID/LawName` tags in houda.go:402-405 | ✓ |
| `_material_key`: `id` -> normalized `downLoadUrl/downloadUrl/url` -> `title` | `houdaMaterialKey` in houda.go:747-758 | ✓, Go 同时接受 `fileUrl/path/materialId/fileId/name` |
| `_make_file_info`: `downLoadUrl/downloadUrl/fileUrl/url`, `title/name`, `file_fmt` | `buildHoudaMaterialEntry` in houda.go:363-384 | ✓ |
| callback query `userId`, `roomId`, `recordId`, `viewerToken` | `resolveHoudaCCCallback` query keys in houda.go:495-500 | ✓ |

## 阻塞步骤

无.
