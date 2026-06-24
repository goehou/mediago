# logic-w1-cross

Scope: `icourse163`, `imooc`, `xuetang`, `zhihuishu`, `dingtalk`, `feishu`, `huke88`, `jianshe99`, `med66`, `kaoyanvip`.

Verification:
- `python3 scripts/verify_full_alignment.py`: PASS. Summary `PASS 91`, `BLOCKED 1`, `PARTIAL 0`, `STUB 0`, `NO_EXTRACT 0`.
- `go build ./...`: PASS.

## icourse163

### Python 流程 (5 步)
1. `Icourse163_App` 先拉已学课程列表和栏目列表, 再根据类型切到 MOOC, Column, 或 Kaoyan flow.
2. `Icourse163_Base` 先注入 `NTESSTUDYSI=srckey` cookie, 然后查订单价格和登录态.
3. `Icourse163_Mooc` 走 `getMocTermDto` -> `getLessonUnitLearnVo` -> `resourceRpcBean.getResourceTokenV2` -> `vod.study.163.com`.
4. `Icourse163_Column` 走 `getMocLessonBaseDtos` -> `getLessonUnitBaseVoByLessonId` -> `getArticleInfoVo`.
5. `Icourse163_Kaoyan` 走 `courseBean.getLastLearnedMocTermDto` / `kaoyanCourseBean.getKyCourseInfoBtStatusVo` / live flow.

### Go 实现
1. 只保留 common `/course/CID-NNN[?tid=MMM]` 入口.
2. 只实现 `getMocTermDto.dwr` + `getLessonUnitLearnVo.dwr` + token/video fallback.
3. 直接拒绝 `/learn/kaopei-...`.
4. 没有 Column, Textbook, App, Package, subtitle, attachment, or course-list fallback handlers.

### 判定
- MISSING_STEP: only the common MOOC flow is ported; Kaoyan, Column, Textbook, App, subtitle, attachment, and learned-course selection branches are absent.

## imooc

### Python 流程 (4 步)
1. `Imooc_Class`, `Imooc_Code`, `Imooc_Free` 都先发 startlearn / ajaxstartlearn 心跳.
2. 取 m3u8/playlist 响应里的 encoded `info` 值.
3. 用 `imooc_decode()` JS 解码, 再跟随解码后的 URL.
4. 收尾时发 endlearn / ajaxendlearn.

### Go 实现
1. 真实发了 startlearn / ajaxstartlearn, 并且在 defer 里发 endlearn / ajaxendlearn.
2. 真实 GET 了 m3u8/playlist 响应.
3. 只接受 free/plain `result` 或原始 m3u8, 付费 encoded URL 直接返回 unsupported error.
4. 没有 JS sandbox, 没有 `imooc_decode` 等价实现, 也没有课程树枚举.

### 判定
- MISSING_STEP: `imooc_decode` 转换链和课程树/课时枚举缺失, 目前只覆盖 free/plain URL 分支.

## xuetang

### Python 流程 (4 步)
1. 从 course 或 learn URL 解析 `sign` 和 `cid`.
2. GET `product/info` 取标题和课程信息.
3. GET `course/chapter`, 再递归展开 `content_data.section_leaf_list.leaf_list`.
4. GET `leaf_info`, 然后 GET `service/playurl/{ccid}` 取最终播放地址.

### Go 实现
1. 同样解析 `sign/cid`.
2. 真实 GET `product/info`, `course/chapter`, `leaf_info`, `service/playurl`.
3. 只做课程主链路, 没有 live, training, 或 fallback 到 `get_course_detail`.
4. `product/info` 失败只做标题兜底, 不回补其他分支.

### 判定
- MISSING_STEP: live / training endpoints and `get_course_detail` fallback are absent; current port is course-only.

## zhihuishu

### Python 流程 (4 步)
1. 普通 course 页面先走 HTML scraping, 解析 course tree.
2. 对每个 video 调 `initVideo`, 拿 `uuid` 和 `lines`.
3. 按 lineID 倒序试 `changeVideoLine`, 取最优 mp4 URL.
4. 另外还有 school / interest / smart / live 等更深的课程家族分支.

### Go 实现
1. 只接受 `videoID=` URL.
2. 只实现 `initVideo` + `changeVideoLine` 这条直视频链.
3. 没有 course-tree scraping, 没有 school / interest / smart / live 入口.
4. 普通 course URL 直接报 “needs HTML scraping that isn't implemented”.

### 判定
- MISSING_STEP: normal course-tree scraping is not implemented; the extractor is still direct-video-only.

## dingtalk

### Python 流程 (5 步)
1. 先按 URL 分出 live-room, group-live-share, preview dentry, 或 alidocs doc 流.
2. live replay 走 LwpClient, 依次 probe live share / live replay / preview dentry / notable record / ai transcribe.
3. `alidocs` 文档流还会用 `api/doc/info`, `api/document/data`, `nt/api/docs/preset/binary`.
4. preset/binary 再补文档下载 URL.
5. 整个 live replay 还要 session-bound token 与 WebSocket / gRPC 风格 RPC.

### Go 实现
1. 能解析 live-room / group-live-share / dentryKey.
2. 只实现 `nt/api/docs/preset` 预览元数据.
3. live replay 直接返回 blocked error, 没有 LWP handshake.
4. `api/doc/info`, `api/document/data`, `nt/api/docs/preset/binary` 都没接.

### 判定
- MISSING_STEP: only the preset preview slice is wired; the live replay handshake and the rest of the alidocs document APIs are absent.

## feishu

### Python 流程 (3 步)
1. `/minutes` 页面抓 `video_url` 并做 unicode-unescape.
2. `/file` 页面抓 `window.SERVER_DATA` 和 `_feishu_preview_url`.
3. `/docx` / `/wiki` 走文档下载流, 不是视频流.

### Go 实现
1. 只实现 `/minutes`.
2. `video_url` 解析和 unicode-unescape 都有.
3. `/file`, `/docx`, `/wiki` 直接返回错误, 没有 preview/doc download.

### 判定
- MISSING_STEP: file preview and doc/wiki document flows are absent.

## huke88

### Python 流程 (5 步)
1. 先用 `_identity-usernew` cookie 做登录态检查.
2. 再拉 course page / current uid / purchased study list.
3. 视频走 `video-play` POST form, 带 `confirm` 重试.
4. 文件走 `video-annex` POST form, 同样带 `confirm` 重试.
5. 课程列表、标题、csrf、课程 ID 提取都在同一条链里闭合.

### Go 实现
1. 同样要求 `_identity-usernew` cookie 并 GET 登录页验证.
2. 真实 GET course page / purchased study list.
3. 真实 POST `video-play` 和 `video-annex`, 也保留 confirm retry.
4. 视频和文件都作为可下载 entry 输出.

### 判定
- ALIGNED: URL, method, auth, parse, and confirm retry all match the source flow.

## jianshe99

### Python 流程 (4 步)
1. 先从视频列表页抓出 lesson / replay 入口.
2. replay payload 从 `liveapi/entry/getReplayInfo` 提取 `liveRoomId`, `accessid`, `recordId`.
3. 再走 CSSLCloud 登录和 VOD 解析.
4. 需要时重写 m3u8 key.

### Go 实现
1. 真实解析 video list / replay anchors.
2. 真实 GET replay payload, 然后交给 `shared.CssLcloudResolvePlayInfo`.
3. 需要时再走 `shared.CssLcloudRewriteM3U8Keys`.
4. 没有内联 CSSLCloud 逻辑, 复用了 shared helper.

### 判定
- ALIGNED: the replay payload shape, csslcloud handoff, and m3u8 rewrite all match the source.

## med66

### Python 流程 (4 步)
1. 先取 member course info.
2. 再取 courseClassWareInfo, 找出录播 / 直播回放入口.
3. replay 页面走 `liveapi/entry/getReplayInfo`.
4. 最后交给 CSSLCloud 登录 / VOD / m3u8 rewrite.

### Go 实现
1. 真实 POST `courseInfo` 和 `courseClassWareInfo`.
2. 真实 GET replay 页面与 `getReplayInfo`.
3. 通过 `shared.CssLcloudResolvePlayInfo` 拿最终播放地址.
4. 需要时再调用 `shared.CssLcloudRewriteM3U8Keys`.

### 判定
- ALIGNED: source URL set, replay payload parsing, and shared csslcloud path are consistent.

## kaoyanvip

### Python 流程 (5 步)
1. 先用 `Pcsite-Token` 走 user info 验证.
2. 再拉 mycourse 列表.
3. 解析 delivery / uuid / outline 树, 找出 video 和 live 节点.
4. 视频走 VOD token -> hls.videocc.net / dpv.videocc.net.
5. live 走 Polyv sign / timestamp / playback API.

### Go 实现
1. 同样先检查 `Pcsite-Token` 和 user info.
2. 同样拉 mycourse 并选 delivery / uuid.
3. 同样展开 outline 树.
4. 同样生成 VOD token, hls/mp4 URL, 和 Polyv live playback URL.

### 判定
- ALIGNED: auth, course selection, outline traversal, VOD token flow, and live playback flow match the Python source.
