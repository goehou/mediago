# htknow 源码对齐对照

## URL 常量

| .cdc.py 行 | htknow.go 行/名 | 一致? |
|---|---|---|
| `Htknow_Base.py:32 referer = 'https://learn.htknow.com'` | `htknow.go:20 refererURL` | ✓ |
| `Htknow_Base.py:257 'https://saas.clientapi.htknow.com/pc_view/learn/list'` | `htknow.go:22 checkCookieURL` | ✓ |
| `Htknow_Course.py:36 course_list_url = 'https://saas.clientapi.htknow.com/learn/list_v2'` | `htknow.go:23 courseListURL` | ✓ |
| `Htknow_Course.py:37 single_url = 'https://saas.clientapi.htknow.com/course/single_detail'` | `htknow.go:24 singleURL` | ✓ |
| `Htknow_Course.py:38 column_url = 'https://saas.clientapi.htknow.com/course/column_course_detail'` | `htknow.go:25 columnURL` | ✓ |
| `Htknow_Course.py:39 series_url = 'https://saas.clientapi.htknow.com/course/series_course_detail'` | `htknow.go:26 seriesURL` | ✓ |
| `Htknow_Course.py:40 live_info_url = 'https://saas.clientapi.htknow.com/live/live_wx/playback_list'` | `htknow.go:27 liveInfoURL` | ✓ |
| `Htknow_Course.py:41 column_info_url = 'https://saas.clientapi.htknow.com/course/column_course_list'` | `htknow.go:28 columnInfoURL` | ✓ |
| `Htknow_Course.py:42 series_info_url = 'https://saas.clientapi.htknow.com/course/series_course_list'` | `htknow.go:29 seriesInfoURL` | ✓ |
| `Htknow_Course.py:43 video_info_url = 'https://saas.clientapi.htknow.com/course/column_play_details'` | `htknow.go:30 videoInfoURL` | ✓ |
| `Htknow_Course.py:44 pc_video_info_url = 'https://saas.clientapi.htknow.com/pc_view/course/column_play_details'` | `htknow.go:31 pcVideoInfoURL` | ✓ |

## 认证 / Header

| 源码方法 (line) | Go 函数 (line) | 一致? |
|---|---|---|
| `_check_cookie` lines 255-267: cookie `token/user/custom_id/base_KEY` + Bearer header | `newCtx` lines 79-95 + `checkCookie` lines 97-107 | ✓ |
| `_check_cookie` lines 269-283: set `authorization/base_KEY/custom_id/user_id/login_user_id/account_list` | `newCtx` lines 83-95 + helpers `accountIDs` | ✓ |

## HTTP 调用

| 源码方法 (line) | Go 函数 (line) | method | 一致? |
|---|---|---:|---|
| `Mooc_Request.request_json` lines 292-306: `requests.post(..., json=data)` | `postJSON` lines 325-340 | POST JSON | ✓ |
| `_check_cookie` line 267 | `checkCookie` line 99 | POST JSON | ✓ |
| `_get_course_list` lines 119-124 | `courseList` lines 136-140 | POST JSON | ✓ |
| `_get_single_info` lines 279-287 | `singleSources` lines 176-184 | POST JSON | ✓ |
| `_get_live_info` lines 320-326 | `liveSources` lines 191-198 | POST JSON | ✓ |
| `_get_column_info` lines 356-364 | `columnSources` lines 209-217 | POST JSON | ✓ |
| `_get_series_info` lines 397-403 | `seriesSources` lines 230-236 | POST JSON | ✓ |
| `_get_product_token` decrypted constants: `video_info_url/pc_video_info_url` | `fetchProductURL` lines 261-287 | POST JSON | ✓ |

## JSON 字段映射

| 源码 key 链 | Go struct/tag 或解析 | 一致? |
|---|---|---|
| cookie `user.id`, `token`, `custom_id`, `base_KEY`, `wechatList/appletList/ksList.child_list.id` | helpers `userIDFromCookie`, `cookieMap`, `accountIDs` | ✓ |
| `learn/list_v2 -> result[] -> product_id/main_product_id/type_desc/title` | `courseList` lines 142-154 | ✓ |
| `single_detail -> result.detail.product_token/pay_content` | `singleSources` lines 183-188 | ✓ |
| `playback_list -> result.list[].video_url/title` | `liveSources` lines 197-203 | ✓ |
| `column_course_list -> result.list[].product_token/pay_content/id/series_id/product_type/title` | `columnSources` + `sourceFromProduct` lines 219-258 | ✓ |
| `series_course_list -> result.list[].article_list[]` | `seriesSources` lines 237-244 | ✓ |
| `column_play_details -> result.article_detail.product_token/pay_content` and `result.detail.*` | `fetchProductURL` lines 277-280 | ✓ |
| `_get_video_url`: product_token base64 JSON `value/iv`, `base_KEY`, AES CBC decrypt, URL starts with `http` | `videoURL` lines 289-322 | ✓ |

## 阻塞步骤

无。`_get_video_url` 在 .cdc.py 第 460 行截断, 已按 R7 读取 `decrypted_full/all_decrypted.json` 中 `Courses/Htknow/Htknow_Course__t343__get_video_url.pyc` 与 `__t360__get_product_token.pyc` 补齐字段链。
