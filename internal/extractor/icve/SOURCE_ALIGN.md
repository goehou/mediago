# icve 源码对齐对照

## URL 常量

| .cdc.py 行 | icve.go 行/名 | 一致? |
|---|---|---|
| Icve_Ai.pyc.1shot.cdc.py:39 `url_title = 'https://ai.icve.com.cn/prod-api/course/courseInfo/getLatestInfoByCourseId?courseId={cid}'` | icve.go:28 `url_title = "https://ai.icve.com.cn/prod-api/course/courseInfo/getLatestInfoByCourseId?courseId=%s"` | ✓ |
| Icve_Ai.pyc.1shot.cdc.py:40 `url_info = 'https://ai.icve.com.cn/prod-api/course/courseDesign/getDesignList?courseInfoId={inf_id}&courseId={cid}'` | icve.go:29 `url_info = "https://ai.icve.com.cn/prod-api/course/courseDesign/getDesignList?courseInfoId=%s&courseId=%s"` | ✓ |
| Icve_Ai.pyc.1shot.cdc.py:41 `url_cell = 'https://ai.icve.com.cn/prod-api/course/courseDesign/getCellList?courseInfoId={inf_id}&courseId={cid}&parentId={parent_id}'` | icve.go:30 `url_cell = "https://ai.icve.com.cn/prod-api/course/courseDesign/getCellList?courseInfoId=%s&courseId=%s&parentId=%s"` | ✓ |
| Icve_Ai.pyc.1shot.cdc.py:42 `url_source_status = 'https://upload.icve.com.cn/{content:}/status'` | icve.go:31 `url_source_status = "https://upload.icve.com.cn/%s/status"` | ✓ |

## HTTP 调用

| 源码方法 (line) | Go 函数 (line) | method | 一致? |
|---|---|---|---|
| Icve_Ai._get_title lines 112-132 | icve.go `loadTitle` lines 108-119 | GET | ✓ |
| Icve_Ai._get_infos lines 141-176 | icve.go `loadItems` lines 122-141 | GET | ✓ |
| Icve_Ai._get_cell_info lines 182-260 | icve.go `loadCellItems` lines 144-152, helpers.go `collectAIItems` lines 268-319 | GET | ✓ |
| Icve_Ai._get_inner_infos lines 266-292 | helpers.go `collectAIItems` lines 268-319 | local JSON tree walk | ✓ |
| Icve_Ai._get_video_url lines 298-361 | icve.go `getVideoURL` lines 220-248, `selectTranscodedURL` lines 262-284 | GET + HEAD-like check via GET | ✓ |
| Icve_Ai._get_file_url lines 367-386 | icve.go `getFileURL` lines 250-260 | local JSON parse | ✓ |
| Icve_Base._video_quality_candidates lines 288-296 | icve.go `videoQualityCandidates` lines 299-308 | local selection | ✓ |
| Icve_Base._select_video_quality lines 302-313 | icve.go `selectVideoQuality` lines 310-318 | local selection | ✓ |

## JSON 字段映射

| 源码 key 链 | Go struct tag / map access | 一致? |
|---|---|---|
| `json.loads(title).get('data',{}).get('id','')` | `aiTitleResp.Data.ID` `json:"id"` | ✓ |
| `json.loads(title).get('data',{}).get('courseName','')` | `aiTitleResp.Data.CourseName` `json:"courseName"` | ✓ |
| `json.loads(title).get('data',{}).get('schoolName','')` | `aiTitleResp.Data.SchoolName` `json:"schoolName"` | ✓ |
| `json.loads(info).get('data', [])` sorted by `sort` | `parseJSONMap` → `listAt(root,"data")` → `sortBySort` | ✓ |
| cell item `id`, `name`, `children`, `fileType`, `fileUrl` | `str(node["id"])`, `str(node["name"])`, `childList`, `node["fileType"]`, `fileInfoText(node["fileUrl"])` | ✓ |
| video info `ossOriUrl`, `ossGenUrl`, `content`, `url` | `data["ossOriUrl"]`, `data["ossGenUrl"]`, `data["content"]`, `data["url"]` | ✓ |
| source status `args`, `type`, quality booleans (`720p`, `480p`, `360p`, `1080p`) | `mapAt(status,"args")`, `status["type"]`, `args[q]` | ✓ |
| file info `ossOriUrl`, fallback `url` | `data["ossOriUrl"]`, `data["url"]` | ✓ |

## 认证与 header

| 源码位置 | Go 对齐 | 一致? |
|---|---|---|
| Icve_Ai.pyc.1shot.cdc.py:78-93 `_check_cookie` / `set_cookie` 均返回 `True` | `Extract` 允许空 CookieJar, 注册 `NeedAuth: false` | ✓ |
| Icve_Base.pyc.1shot.cdc.py:104-112 `Sec-Fetch-*`, `Sec-Ch-Ua-*`, `Referer`, `cookie` | `newCtx` lines 94-103 同名 header | ✓ |

## 返回结构

| 源码行为 | Go 行为 | 一致? |
|---|---|---|
| `_download_course` 遍历 `video_list` / `file_list` 和嵌套章节 | `mediaFromItems` 返回 `MediaInfo.Entries`, 每个 entry 带一个 stream | ✓ |
| `mode == ONLY_PDF` 时 `_download_video_list` 跳过视频 | `mediaFromItems` 在 `ONLY_PDF` 下跳过 video entry | ✓ |

## 阻塞步骤

无。
