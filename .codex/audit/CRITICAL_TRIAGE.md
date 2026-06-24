# CRITICAL Triage

## MUST FIX (real code issues)

### 1. douyin nil panic (w1)
- `douyin.go:92-94,115-116,138-140,291-295` — http.NewRequest error unchecked, immediate dereference
- Fix: check err before using req

### 2. gaotu multi-domain routing (w2)
- gaotu.go:27 registers gaotu100.com/gtgz.cn/naiyouxuexi.com patterns but always calls api.gaotu.cn
- Source subclasses use brand-specific hosts (Gaotu_Tutu→gaotu100, Gaotu_Gaozhong→gtgz, Gaotu_Suyang→naiyouxuexi)
- Fix: detect domain from URL, route to correct API host

### 3. htknow answer endpoints missing (w2)
- Source has answer_tag_url, answer_num_url, answer_list_url, answer_create_paper_url (Htknow_Course.py:45-48)
- Go constants stop at video/detail
- Fix: add answer endpoints to constants

### 4. htknow HTML items dropped (w2)
- mediaFromSources only appends when src.url != "" — drops pure 图文 entries
- Fix: keep HTML content entries as downloadable items

### 5. qlchat dead Qianliao train flow (w4)
- Qianliao train endpoints declared but never called
- Fix: implement or mark explicitly as not-implemented subflow

## EXPECTED BLOCKED (not bugs, by design)

### dingtalk live replay (w1)
- LWP WebSocket protocol — 3000+ lines, impossible without protocol stack
- Already returns blocked error. NOT a bug.

### imooc_decode (w3)
- JS sandbox required — already returns blocked error for paid content
- NOT a bug.

### icourse163 kaopei/column/textbook (w3)
- Sub-site flows not implemented, explicitly rejected
- Acceptable for v2 (main /course/ flow works)

### feishu /file /docx /wiki (w2)
- Document download flows, not video — returns clear error
- Acceptable (video /minutes path works)

### zhihuishu course-tree (w6)
- HTML scraping needed for course tree, only videoID URL works
- Known limitation, documented

### Aliyun VOD/MTS encrypted flows (w5: wangxiao233, wowtiku, xiaoeapp, xiaoetech)
- DRM/STS license flows need full Aliyun SDK integration
- Complex, marked as partial

## FALSE POSITIVES

### bilibili (w1)
- Auditor didn't know bilibili.go is public-video flow, cheese.go is pugv flow
- Both exist, both correct. NOT a bug.

### douyin "no source" (w1)
- Source is in ~/code/clis/douyin-dl/src/ not Courses/ dir
- NOT fabricated.
