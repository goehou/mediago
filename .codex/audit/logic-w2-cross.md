# logic-w2-cross

Scope: `ahu`, `cctalk`, `cto51`, `dongao`, `fenbi`, `gaodun`, `hqwx`, `huatu`, `wangxiao`, `mashibing`, `plaso`.

Review boundary: 以 `~/code/xwz-downloader-source-release/decompiled_full/Mooc/Courses/<Site>/*_Course.pyc.1shot.cdc.py` 为主; 当 cdc 在反编译截断处缺少后半段下载/解密逻辑时, 使用同目录 `.das` 反汇编补足可执行分支. Go 侧只审计 `internal/extractor/<site>/*.go` 的当前实现.

## ahu
### Python 流程 (6 步)
1. `Ahu_Base._check_cookie` 依赖 `phpsessid`/`laravel_session`, GET `/center/mycourse.html?_=...`, 要求仍在 `/center/mycourse.html` 且页面包含退出登录和课程块.
2. `_get_cid` 从 URL 提取 `courseId`, 否则分页 GET `/center/mycourse.html?page=N` 枚举 `div.yxg-mc-student` 并选择课程.
3. `_get_detail_soup` GET `course_info_url`; `_parse_course_videos` 从 `/video/videoplay.html?...lessonId=...` 组装视频树.
4. `_get_infos` 同时填充视频 `infos` 和 `_source_info`; `_source_info` 来自 `handoutsList` JSON 与课程页面文件链接.
5. `_get_play_info` GET 播放页后返回 Aliyun `(videoId, playAuth)`, Baijiayun `(hlsToken, playId)`, 或 direct `m3u8/video` URL.
6. `_download_video` 对 Aliyun 调 `GetPlayInfo` 并处理 `MtsHlsUriToken`/加密 HLS key; Baijiayun 分支走回放下载; direct 分支走 m3u8/mp4 下载.
### Go 实现
1. `Extract` 只要求 cookie jar, 从 URL 解析 `courseId`/`lessonId`; 未执行 Python 的登录探测和课程列表选择.
2. GET `course_info_url`, `parseLessons` 只提取 `/video/videoplay.html` 课时链接.
3. `resolveLesson` GET 播放页, 支持 direct `videoSrc|m3u8_url` 和 Aliyun `aliyunVideoId|vodVideoId|aliyunVid` + `playAuth`.
4. `requestAliyunPlayInfo` 解码 `playAuth`, HMAC-SHA1 签 `GetPlayInfo`, 按清晰度排序返回 `PlayURL`.
5. 未覆盖 Baijiayun `hlsToken/playId`, Aliyun 加密 HLS token/key 下载路径, 以及课程 handout/file 输出.
### 判定
- MISSING_STEP: Go 只覆盖 lesson/direct/Aliyun 基础播放链, 缺少 Python 决定性下载分支 Baijiayun, 加密 HLS key/token, 以及 `_source_info` 文件链.

## cctalk
### Python 流程 (7 步)
1. `Cctalk_Base` 规范化 `ClubAuth`/`access_token` 等 cookie/token, API 统一带 `pcweb` 应用头.
2. `_get_group_video_list`, `_get_series_content_list`, `_get_series_all_lesson_list`, `_get_series_structs`, `_get_course_structs` 组合获取班级, 系列和课程结构.
3. `_get_video_play_info` 先 GET `/video/detail` v1.1/v1.2, 再 POST `/video/play` v1.1/v1.2 取播放元数据.
4. `_get_subscribe_courses` 在 URL 无明确课程时调用移动端 `user/my_group_list` 做已购课程枚举.
5. `_iter_material_candidates`, `_build_file_info`, `_get_article_detail` 构造资料, 文章和文件节点.
6. 课件播放会提取 `coursewareId`, `userSign`, `videoUrl/fileUrl`, 并解析 OCS 资源.
7. `_prefer_v55_*`, `_build_cctalk_v55_m3u8_media_info`, `_download_video` 处理 v55/OCS/board/AES segment 播放与下载.
### Go 实现
1. `Extract` 要求 cookies, 解析 group/series/course/video ID; 有 video ID 时直接取播放信息.
2. `getSeriesStructs`, `getGroupVideoList`, `getCourseStructs` 读取系列, 班级和课程结构.
3. `getVideoPlayInfo` 对齐调用 `/video/detail` 和 `/video/play` 的 v1.1/v1.2 组合.
4. `buildEntries` 递归结构体, 对没有直接 URL 但有 videoID 的节点补取播放信息.
5. `findMediaURL` 只接受现成 `.m3u8/.mp4` 类 URL; 未解析 OCS/v55/userSign/board, 也未输出资料文章文件.
### 判定
- MISSING_STEP: API 结构和普通视频 play/detail 链基本对齐, 但 Go 缺少 Python 的已购课程枚举, OCS/v55/AES/board 下载链, 以及 article/file/material 分支.

## cto51
### Python 流程 (7 步)
1. Course 常量覆盖 lesson list, lesson file list, course material/file list, vod play auth, QCloud play API.
2. `_get_infos` 组合课程目录, 训练营/阶段/直播/课件列表, 并生成视频与资料两类节点.
3. `_fetch_lesson_file_list_payloads` 和 `_fetch_course_file_list_payloads` 拉取课时资料与课程资料.
4. `_extract_aliplayparam` 从播放页提取 Aliyun 参数.
5. `_derive_51cto_sec_data`, `_decrypt_51cto_cbc_text`, `_decode_aliyun_play_auth` 还原 Aliyun playAuth/secData.
6. `_request_aliyun_play_info_by_rand`, `_request_aliyun_play_info_legacy`, `_request_aliyun_play_info` 进入 Aliyun `PlayInfoList` 链.
7. `_request_qcloud_play_info`, `_download_video_item`, `_download_file_list` 分别处理 QCloud 视频与文件下载.
### Go 实现
1. `Extract` 解析 course/lesson/training/order 参数, 无参数时查 mycourses.
2. `resolveCourse` 请求 lesson list, lesson file list, course file list; fallback 从 HTML 提取 lesson 链接.
3. `resolveTraining` 并发式请求训练营 stage/course/info/live/file/next API 并递归抽取媒体.
4. `resolveLesson` GET 播放页, 尝试 direct media, `parseAliPlayParam`, 再 fallback `vod_play_auth_api`.
5. `resolveAuth` 只覆盖 QCloud `app_id/file_id/psign` 和 direct `play_url`; 文件节点被 `mediaFromText` 的 `.m3u8/.mp4` 过滤规则大量丢弃.
### 判定
- MISSING_STEP: Go 覆盖 QCloud/direct 和部分列表 API, 但缺少 Python 的 51CTO CBC/secData/Aliyun GetPlayInfo 解密链, 且资料下载链未按 `_download_file_list` 对齐.

## dongao
### Python 流程 (6 步)
1. `_request_stage_list`, `_request_detail_infos`, `_get_catalog_html` 拉取阶段, 明细和目录 HTML.
2. `_get_infos` 解析课程目录和 lecture 节点.
3. `_get_lecture_page` 用 `playerType` 请求 lecture 播放页.
4. `_extract_player_fields` 提取 `listenParam`/player 字段.
5. `_aes_cbc_decrypt` 使用 Dongao 固定 AES key/iv; `_pick_video_source` 解密 `source/url` 并选择真实视频源.
6. `_download_video` 对 m3u8 构造 signed m3u8/meta, 否则走普通视频下载.
### Go 实现
1. `Extract` 解析 course/lecture ID 并要求 cookies.
2. `resolveCourse` GET catalog, 从 HTML/JSON 递归寻找 media/lecture 链接.
3. `requestCourseAPIs` POST stage/detail/live/catalog 探针接口.
4. `resolveLecture` POST `lecture_url` 带 `playerType=h5`, fallback GET.
5. `findMediaInText` 只递归寻找 direct `url/path/playUrl/m3u8/mp4`, 无 AES 解密和 signed m3u8 构造.
### 判定
- MISALIGNED: Go 的课程 API 顺序部分覆盖, 但视频真实地址在 Python 中依赖 AES-CBC player 字段解密和 signed m3u8 处理, Go 只做明文 URL 扫描.

## fenbi
### Python 流程 (7 步)
1. `Fenbi_Base._check_cookie` 调 `login_check_url` 校验登录; Course 侧还区分不同 course prefix.
2. `_get_course_prefix_list` 调 `course_list_url`; `_get_visible_lectures` 分页枚举可见讲次.
3. `_get_lecture_detail`, `_get_lecture_summary`, `_get_episode_nodes`, `_get_episode_detail` 组成 lecture/episode 树.
4. `_resolve_material_url` 调 `material_url`, 失败再试 `vertical_material_url` 获取资料.
5. `_get_infos` 同时产出视频与 `source_info` 文件节点.
6. `_get_video_url` 对 `media_meta_url` 尝试多组 `(biz_type,biz_id)` 与 `content_id` 参数组合.
7. `_download_video` 下载 m3u8/mp4, 并在需要时走 replay-board fallback.
### Go 实现
1. `Extract` 要求 cookies, 解析 lecture/episode ID; 未提供课程 ID 时直接报错.
2. `checkLogin` 请求 login/ke 检查, 但最终失败也返回 nil, 不阻断.
3. `resolveLecture` 调 lecture detail, lecture summary, episode_nodes, 并递归收集 episode.
4. `resolveEpisode` 调 episode detail 和一次 `media_meta_url`, 递归挑第一个视频 URL.
5. 未调用 course list/visible lectures/material/vertical material, 也未实现多组 media_meta 参数和 replay-board 分支.
### 判定
- MISSING_STEP: Go 覆盖单 lecture/episode 的主视频 Happy Path, 缺少 Python 的课程枚举, 资料 URL 解析, media_meta 多参数回退, 登录失败判定和 replay-board 兜底.

## gaodun
### Python 流程 (7 步)
1. `_get_course_list` 调 `course_url` 读取 `result.courseList` 并选择课程.
2. `_get_pc_token` GET `pc_token_url`, `_get_pe_token` GET `pe_token_url`, 保存直播/回放所需 token.
3. `_get_infos` 从 `result.syllabus.children` 构造章节和课时.
4. `_get_chapter_info` 和 `_get_source_info` 拉取 syllabus, source category/gradation 数据.
5. `_get_video_url` 调 `video_play_url` 并读取 `result.path`.
6. `_get_live_url` 先调 `live_token_url`, 再调 `live_play_url`; `_get_live_old_url` 读取 `result.playUrls[].playUrl`.
7. `_get_file_url` 调 `file_url`; `_download_video` 在 live/old/live_token 分支前确保 pc/pe token 已获取.
### Go 实现
1. `Extract` 要求 cookies 并从 URL 解析 course/video/record/token; 无明确 ID 时报错.
2. `resolveCourse` 请求 info, gradation, source category/source gradation, syllabus/glive 相关接口.
3. `resolveDirect` 对 record 调 `live_old_url`, 对 token 调 `live_token_url`, 对 video 轮询 FHD/HD/SD 的 video/live play URL.
4. `pickMedia` 递归提取 direct path/playUrl.
5. 未调用 `course_url` 做课程选择, 未调用 pc/pe token API, 文件下载也未按 `_get_file_url` 的资源 ID 链精确解析.
### 判定
- MISSING_STEP: Go 的点播/直播 URL API 覆盖不完整; Python 要求的课程列表入口, pc/pe token 前置步骤和 file_url 资料链缺失.

## hqwx
### Python 流程 (6 步)
1. Base 从 cookies 解析 passport 与 `edu24ol-token`, `base_params` 统一拼接 `domainId`, `token`, `puid` 等参数.
2. `_adminapi_header` 和通用 JSON GET/POST 统一处理主站, admin API 和 open API 头.
3. `_request_course_list`, `_request_stages`, `_request_stage_tasks`, `_request_schedules`, `_request_lessons`, `_request_course_detail` 组成普通课程/阶段/任务/课时流.
4. `_request_plan_categories`, `_request_plan_groups`, `_request_plan_lessons` 覆盖 study plan 流.
5. `_request_video_resource`, `_request_live_playback_resource`, `_request_subtitle_url`, `_request_last_video_log` 获取视频, 回放, 字幕和学习记录.
6. `_get_infos_open_course` 与 v2 lesson list 分支覆盖公开课/新版课时.
### Go 实现
1. `newCtx` 从 cookie 还原 passport, token 和 `edu24ol-token`; `baseParams` 与 Python 参数集合一致.
2. `loadJSONGet`, `loadJSONPost`, `adminAPIHeaders` 统一封装主站/admin/open API 请求.
3. `api.go` 实现课程列表, stage, stage task, schedule, lesson, course detail, plan category/group/lesson, video/live/subtitle/log/open lesson 等接口.
4. `flows.go` 按 TYPE_STAGE_TASK, TYPE_SCHEDULE_LESSON, TYPE_OPEN_COURSE, TYPE_STUDY_PLAN 组装入口流.
5. `media.go` 按 Python 字段优先级选择 fhd/hd/sd m3u8/url 并附加资料和字幕.
6. 错误处理按 code/result/message 归一, 无明显缺失的下载分支.
### 判定
- ALIGNED: Go 请求顺序, token/header/base params, 普通课/计划课/公开课和 video/live/subtitle/material 解析均与 Python 主流程对齐.

## huatu
### Python 流程 (6 步)
1. Base 的 `_apply_token_headers`, `_get_cookie_token`, `_normalize_cookie_token_aliases` 统一 token/cookie 别名与请求头.
2. `_get_course_list` 拉取 `my_course_url`; `_get_syllabus_items` 按 `goodsNum`, `level`, `stageId`, `modularId` 分页/分层获取 syllabus.
3. `_get_infos` 递归 syllabus, 同时收集视频和文件节点.
4. `_get_video_source` 调 `player_url`, 区分 Baijiayun 与腾讯 VOD.
5. 腾讯 VOD 分支提取 `drmToken`, 选择最高 `BANDWIDTH` media playlist, `_append_token` 和 `_rewrite_m3u8_text` 重写 key/token.
6. Baijiayun 分支由 `_extract_baijiayun_play_source` 和后续下载函数处理; `_download_one_file` 处理资料下载.
### Go 实现
1. `newCtx` 规范化 token aliases 和请求头; `getJSON`/`successCode` 对齐 Python JSON 成功判定.
2. `courseList`, `syllabusItems`, `collectItems` 覆盖 my_course 和 goodsNum/level/stage/modular syllabus 递归.
3. `collectItems` 同时追加 video 与 file/material entries.
4. `videoSource` 调 `player_url`, 支持 Baijiayun VOD/Playback shared resolver.
5. 腾讯 VOD 分支调用 `vod_info_url`, 提取 `drmToken`, 拉最终 m3u8, 选择最高 `BANDWIDTH` 并重写 `EXT-X-KEY` URI.
6. 文件 URL 和标题解析与 Python `_download_one_file` 输入保持一致.
### 判定
- ALIGNED: Go 覆盖 Python 的 token/header, course/syllabus, Baijiayun, 腾讯 DRM m3u8 rewrite 和资料节点流程.

## wangxiao
### Python 流程 (7 步)
1. Base `_check_cookie` 验证登录态; Course 侧可从订单/课程入口选择课程.
2. `_get_user_catalog_groups` 调 `ProductsDirectory` 与 `GetClasshours` 构造用户目录.
3. `_get_ke_catalog_groups` 调 ke catalog, 解析视频, 直播, 文件和 handout 字段.
4. `_get_infos` 合并 user/ke catalog 产出视频和资料.
5. `_get_lesson_payload` 访问 lesson/play/player 页面, 提取 BokeCC `siteid`, `cc_vid`, 以及 `DownHandOut|down.aspx` handout URL.
6. `_get_video_url` 调 BokeCC `getvideofile`; `_get_m3u8_text` 获取 m3u8 文本.
7. `_resolve_file_resource`, `_download_one_file` 解析 `file_html`, `file_url`, handout headers, 文件名和下载 URL.
### Go 实现
1. `Extract` 请求 seed page, 做一次 login page 检查, 从 URL/pageData/href 提取 lesson ref.
2. `refsFromKeCatalog` 只对 `setmealId` 调 `skuSingleContent` 并构造视频 ref.
3. `resolveRef` 请求 lesson/play/player 页面并提取 BokeCC `siteid/vid`.
4. 调 `shared.BokeCCResolve` 生成视频 media.
5. 常量中存在 `urlDirectory`/`urlClasshours`, 但当前流程未调用; handout/file/html 下载分支未输出.
### 判定
- MISSING_STEP: Go 覆盖 BokeCC 视频解析, 但缺少 Python 的 ProductsDirectory/GetClasshours 用户目录, `DownHandOut/down.aspx` handout, `file_html/file_url` 资料解析和下载链.

## mashibing
### Python 流程 (7 步)
1. Base 从 cookie 设置 token header, GET `/uaa/user` 校验登录.
2. `_get_course_list` 读取 `ownCourse` 和 `ownPackageList`; `_get_price` 补价格/权限信息.
3. 课程详情流读取 systemCourse/course, courseWeb/pc, chapter/section 列表和文档接口.
4. 文件资料包括 dataUrl, gitUrl, netdiskUrl, fynoteUrl, downloadUrl, attachment/file URL.
5. Polyv secure 流: `_format_polyv_vid`, `_build_polyv_request_headers`, `_decrypt_polyv_secure_info`, `_decrypt_polyv_key`.
6. PDX 流: `_build_polyv_pid`, `_build_polyv_pdx_key_url`, `_build_polyv_token_key_url`, `_decrypt_polyv_pdx_text`.
7. `_build_polyv_pdx_info`, `_build_polyv_pdx_m3u8_meta` 生成 PDX m3u8 meta/key, 再进入下载.
### Go 实现
1. `mashibingBuildSession` 从 cookies 设置 token, GET `/uaa/user` 并要求 code 200.
2. course list 覆盖 ownCourse/ownPackageList 分页; 详情覆盖 `/systemCourse/course/{cid}` 与 `/courseWeb/{cid}/pc`.
3. `mashibingBuildItems` 递归 chapter/section, 收集 Polyv videoId 和多类文件 URL; 还读取 document list/info.
4. `mashibingPolyvInfo` 解 Polyv secure JSON; `mashibingPolyvDecode` 实现 AES-CBC MD5 key/iv 解密.
5. `playSafe` 和 `shared.PolyvRewriteM3U8Keys` 覆盖普通 Polyv HLS key 重写.
6. `mashibingBuildPolyvPDXURL` 只能构造 PDX 入口 URL, 未完成 Python 的 PDX key/token 文本解密和 m3u8 meta 构建.
### 判定
- MISALIGNED: 登录, 课程, 文档和普通 Polyv secure/playSafe 链对齐; PDX 加密播放链仍缺少 Python 的 PDX decrypt/key/meta 逻辑.

## plaso
### Python 流程 (7 步)
1. `Plaso_Base._plaso_request_json` 统一请求 JSON; `_check_cookie` 校验登录.
2. `set_mode` 控制 video, board, courseware, `download_player_html` 等下载模式.
3. `_get_history_list`, course/share/package/file/info API 组合获取课程包和文件.
4. Aliyun 分支调用 `m3u8_url`; Polyv 分支调用 polyv sign/video/m3u8 sign 相关接口.
5. `download_player_video` 使用 `player_url` 和 `plaso_player_url_encrypt` 生成离线 player HTML.
6. 同目录 `Aiwenyun_Course` 与 `Jhpy_Course` 复用同类 API, 但 host 分别为 `www.aiwenyun.cn` 和 `jhpy.plaso.cn`.
7. 文件, board, courseware 分支按模式分别下载或生成资源.
### Go 实现
1. `Extract` 解析 `sfid/fileId/id/packageId`; 无明确 ID 时查 course/list/package/history.
2. `fetchPackageFiles` 请求 package, info, dir_info; `fetchShareOrFile` 请求 share/file/file_info.
3. `resolveFile` 从 direct URL, Ali getPlayInfo, Polyv, 或合成 location 文件 URL 中选媒体.
4. `fetchAliPlayURL` 调 `m3u8_url`; `fetchPolyvURL` 覆盖 polyv sign, secure, video, m3u8 sign fallback.
5. URL pattern 和常量仅覆盖 `plaso.cn`; 未处理 Aiwenyun/Jhpy host, board/courseware/player HTML 输出.
### 判定
- MISSING_STEP: 主 Plaso 媒体 API 大体覆盖, 但 Python 的 player HTML 加密产物, board/courseware 模式, 以及 Aiwenyun/Jhpy 子站 host 路由未对齐.
