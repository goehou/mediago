# med66 源码对齐对照

## URL 常量

| .cdc.py 行 | med66.go 行/名 | 一致? |
|---|---|---|
| `Med66_Config.py:36 LOGIN_URL = 'https://www.med66.com/OtherItem/loginAgain/index.shtml'` | `med66.go:17 LOGIN_URL` | ✓ |
| `Med66_Config.py:37 MEMBER_HOME_URL = 'https://member.med66.com/homes/mycourse'` | `med66.go:18 MEMBER_HOME_URL` | ✓ |
| `Med66_Config.py:38 COURSE_INFO_URL = 'https://member.med66.com/homes/mycourse/courseInfo'` | `med66.go:19 COURSE_INFO_URL` | ✓ |
| `Med66_Config.py:39 COURSEWARE_INFO_URL = 'https://member.med66.com/homes/course/courseClassWareInfo'` | `med66.go:20 COURSEWARE_INFO_URL` | ✓ |
| `Med66_Config.py:41 ELEARNING_HOME_URL = 'https://elearning.med66.com/'` | `med66.go:21 ELEARNING_HOME_URL` | ✓ |
| `Med66_Course.py:41 live_replay_referer = 'https://live.cdeledu.com/'` | `med66.go:23 LIVE_REFERER_URL` | ✓ |
| `Med66_Course.py:42 live_replay_info_url = 'https://live.cdeledu.com/liveapi/entry/getReplayInfo'` | `med66.go:22 LIVE_REPLAY_INFO_URL` | ✓ |

## HTTP 调用

| 源码方法 (line) | Go 函数 (line) | method | 一致? |
|---|---|---:|---|
| `Med66_Course._get_course_list` line 137: `_post_form_json(COURSE_INFO_URL,{})` | `fetchCourse` line 120: `PostForm(COURSE_INFO_URL,{})` | POST form | ✓ |
| `Med66_Course._get_coursewares` line 446: `_post_form_json(COURSEWARE_INFO_URL,payload)` | `fetchCoursewares` line 156: `PostForm(COURSEWARE_INFO_URL,form)` | POST form | ✓ |
| `Med66_Course._resolve_live_replay_payload` lines 1337,1347: GET play URL then `getReplayInfo` JSON | `resolveReplayPayload` lines 263,274 | GET + GET JSON | ✓ |
| `Med66_Course._login_live_replay_cc` lines 1399-1411 and `_get_live_replay_context` line 1132 | `resolveReplayEntry` lines 199-208 via `shared.CssLcloudResolvePlayInfo` | CSSL helper | ✓ |
| `Med66_Course._prepare_live_replay_m3u8_text` line 1622 | `resolveReplayEntry` lines 236-239 via `shared.CssLcloudRewriteM3U8Keys` | GET m3u8 + key rewrite | ✓ |

## JSON 字段映射

| 源码 key 链 | Go struct/tag 或解析 | 一致? |
|---|---|---|
| `COURSE_INFO_URL -> data/result list -> courseId, eduSubjectId, classType, classId, linkedCourseIds` | `collectMaps` + `med66Course` fields in `fetchCourse` | ✓ |
| `COURSEWARE_INFO_URL -> homeCwareList/homeWareList/courseWareList/wareList` | `fetchCoursewares` keys list | ✓ |
| `html onclick -> goToLive('...')` | `goToLiveRe` same regex body | ✓ |
| `payload['data']['replay'] or payload['replay']` | `liveReplayPayload.Data.Replay` / `Replay` | ✓ |
| `replay.liveRoomId/liveId/accessid/recordId/accesskey, token` | `liveReplayReplay` tags + `Token` | ✓ |
| `csslcloud datas.sessionId, data.vod_info.video/audio` | `shared.CssLcloudResolvePlayInfo` | ✓ |

## 阻塞步骤

无。CSSL replay 登录, vod 解析和 m3u8 key 处理按任务要求复用 `internal/extractor/shared/csslcloud.go`。
