# CCTalk 源码对齐

## URL 常量

| .cdc.py 行 | cctalk.go 行/名 | 一致? |
|---|---|---|
| `Cctalk_Config.pyc.1shot.cdc.py:55-62` `CCTALK_BASE_URL`, `CCTALK_CONTENT_API_V1/V11/V12`, `CCTALK_PCWEB_KEY`, `CCTALK_TENANT_ID`, `CCTALK_USER_AGENT` | `cctalk.go:15-21` | ✓ |
| `Cctalk_Course.pyc.1shot.cdc.py:1901-1903` `my_group_list`, `mycourse`, mobile origin | `cctalk.go:23-25` | ✓ |

## HTTP 调用

| 源码方法 | Go 函数 | method | 一致? |
|---|---|---|---|
| `_api_url` / `_request_api` (`Cctalk_Course.pyc.1shot.cdc.py:984-1010`) | `apiURL` / `requestAPI` (`cctalk.go:104-125`) | GET/POST | ✓ |
| `_get_course_structs` (`Cctalk_Course.pyc.1shot.cdc.py:1616-1630`) | `getCourseStructs` (`cctalk.go:145-156`) | GET | ✓ |
| `_get_series_all_lesson_list` / `_get_series_content_list` (`Cctalk_Course.pyc.1shot.cdc.py:1313-1412`) | `getSeriesStructs` (`cctalk.go:158-177`) | GET | ✓ |
| `_get_group_video_list` (`Cctalk_Course.pyc.1shot.cdc.py:1226-1249`) | `getGroupVideoList` (`cctalk.go:179-184`) | GET | ✓ |
| `_get_video_play_info` (`Cctalk_Course.pyc.1shot.cdc.py:1782-1859`) | `getVideoPlayInfo` (`cctalk.go:186-197`) | GET + POST | ✓ |

## JSON 字段映射

| 源码 key 链 | Go 解析点 | 一致? |
|---|---|---|
| `_extract_data`: `data` / `Data` / `result` / `Result` | `extractData` (`cctalk.go:254-263`) | ✓ |
| `_extract_list`: `items`, `list`, `lessonList`, `videoList`, `contentList` | `extractList` (`cctalk.go:265-277`) | ✓ |
| `_walk_nodes`: `children`, `lessons`, `lessonList`, `items`, `list`, `contentList`, `videoList`, `mediaList`, `playList` | `walkMaps` (`cctalk.go:279-297`) | ✓ |
| `_build_video_info`: `videoUrl`, `playUrl`, `m3u8Url`, `hlsUrl`, `mediaUrl`, `mediaURL`, `mp4URL`, `url` | `findMediaURL` (`cctalk.go:299-324`) | ✓ |
| `_node_title`: `lessonName`, `videoName`, `contentName`, `title`, `name`, `subject` | `mediaFromMap` (`cctalk.go:216-230`) | ✓ |

## 阻塞步骤

无。
