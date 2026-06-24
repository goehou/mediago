# worker-6 extended goal: 完成原 3 站 + 新增以下 zlketang 站

## 你已经在做的 3 站 (继续完成)
按原计划提交.

## 新增任务清单 (按顺序做)
yangcong yikaobang yixiaoerguo yizhiknow youdao youyuan youzan zhaozhao zhengbao zlketang

## 规则不变 (再次强调)
1. 每站 `internal/extractor/<site>/<site>.go` 必须真实调 HTTP + 真实解析响应. 禁止 stub 和假成功.
2. URL 常量原样照抄 `~/code/xwz-downloader-source-release/decompiled_full/Mooc/Courses/<SourceDir>/` 的 .cdc.py.
3. JSON 路径照抄源码 .get('xxx') 链.
4. csslcloud 站调 shared.CssLcloudResolvePlayInfo, polyv 调 shared.PolyvResolveSecure, bokecc 调 shared.BokeCCResolve, baijiayun 调 shared.BaijiayunResolveVOD/BaijiayunResolvePlayback. 不要重写签名.
5. 每站写 SOURCE_ALIGN.md (URL/HTTP/JSON 三栏对照表).
6. 每站独立 commit.
7. 站点源码本身只有 home URL 无 API 样本的 (yikaobang/cnmooc 等), 在源码里有明确标识时, 在 SOURCE_ALIGN.md 标 BLOCKED + 原因, Extract() 返回 "blocked: needs upstream API samples".

## 自检审计 (每站完成后必做)
- `cd $WORKTREE && go build ./... && go vet ./internal/extractor/<site>/...`
- `python3 scripts/verify_full_alignment.py | grep <site>` 必须显示 PASS 或 BLOCKED, 不允许 STUB
- 跑 `go test ./internal/extractor/<site>/...` (如果你写了 _test.go)

## 交叉审计 (你完成全部本人任务后)
完成自己全部站后, 切到主仓 `~/code/medigo`, 用 `git log work/v2-batch1-w<邻居>` 看邻居的提交.
随机抽 2 站, 读那 2 站的 .cdc.py + 邻居的 Go 代码 + SOURCE_ALIGN.md, 找差异.
发现问题: 在 `~/code/medigo/.codex/audit/<your-w>-reviewed-<neighbor-w>.md` 写问题清单 (issue + line + fix suggestion).

邻居映射: w1↔w2, w3↔w4, w5↔w6.

## 全部完成后
`git push -u origin work/v2-batch1-w6`
然后在 pane 里写 "DONE: 14 sites committed + audit report at .codex/audit/w6-reviewed-w?.md"

## 你的工作目录
/home/sophomores/code/medigo-w6 (branch work/v2-batch1-w6)

立即继续, 不要等确认.
