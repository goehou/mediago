# speiyou 源码对齐对照

## URL 常量

| .cdc.py 行 | speiyou.go 行/名 | 一致? |
|---|---|---|
| Speiyou_Base.py:30 `referer = 'https://speiyou.cn/'` | `referer` | ✓ |
| Speiyou_Base.py:31 `check_api = 'https://course-api-online.speiyou.com/course/v1/student/course/subject-list?stuId={}'` | `subject_api` | ✓ |
| Speiyou_Course.py:32-35 subject/course/chapter/live APIs | `subject_api`, `course_list_api`, `chapter_list_api`, `live_list_api` | ✓ |
| Speiyou_Course.py:36 `auth_api = 'https://classroom-api-online.speiyou.com/classroom/basic/v2/init/auth?resVer=1.1&classroomMode=playback'` | `auth_api` | ✓ |
| Speiyou_Course.py:37 `video_api = 'https://classroom-api-online.speiyou.com/playback/v1/video/init'` | `video_api` | ✓ |

## HTTP 调用

| 源码方法 (line) | Go 函数 | method | 一致? |
|---|---|---|---|
| `_request_json` 57-80 | `requestJSON` | GET | ✓ |
| `_get_subjects` 332-340 | `Extract` subject probe | GET | ✓ |
| `_get_live_list` 241-271 | `fetchLiveList` | GET | ✓ |
| `_get_legacy_course_list` 345-389 | `fetchCourseAndLessons` | GET | ✓ |
| `_get_legacy_live_list` 555-574 | `fetchLegacyLessons` | GET | ✓ |
| `_merge_auth_info` 702-771 | `resolveVideo` / `mergeAuthInfo` | GET | ✓ |
| `_get_video_url` 777-799 | `resolveVideo` | GET | ✓ |

## JSON 字段映射

| 源码 key 链 | Go struct / map 访问 | 一致? |
|---|---|---|
| `data/list/result/records` | `jsonToMaps` | ✓ |
| `stdCourseId/course_id/courseId` | `courseKey` | ✓ |
| `liveId/live_id` | `lessonKey` | ✓ |
| `initData.live/course/classInfo/teacher` | `mergeAuthInfo` | ✓ |
| `videoUrls[]` / `videoUrl` | `resolveVideo` | ✓ |
| playback headers `liveType`, `stdClassId`, `stdSubject`, `stdCourseId`, `liveId`, `stuId`, `token` | `playbackHeaders` | ✓ |

## 阻塞步骤

无.
