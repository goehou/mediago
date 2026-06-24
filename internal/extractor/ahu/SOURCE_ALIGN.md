# Ahu 源码对齐

## URL 常量

| .cdc.py / .das 行 | ahu.go 行/名 | 一致? |
|---|---|---|
| `Ahu_Course.pyc.1shot.cdc.py:38-40` | `ahu.go:25-27` `course_list_url`, `course_info_url`, `video_play_url` | ✓ |

## HTTP 调用

| 源码方法 | Go 函数 | method | 一致? |
|---|---|---|---|
| `_get_detail_soup` / `_request_text` (`Ahu_Course.pyc.1shot.cdc.py:416,337`) | `Extract` (`ahu.go:80-83`) | GET | ✓ |
| `_get_play_info` / `_request_text` (`Ahu_Course.pyc.1shot.cdc.py:868,337`) | `resolveLesson` (`ahu.go:138-145`) | GET | ✓ |
| `_request_aliyun_play_info_by_rand` / `_request_aliyun_play_info_legacy` (`Ahu_Course.pyc.1shot.das:10137-11096`) | `requestAliyunPlayInfo` (`ahu.go:191-242`) | GET | ✓ |

## JSON 字段映射

| 源码 key 链 | Go struct tag / 解析点 | 一致? |
|---|---|---|
| `PlayInfoList -> PlayInfo -> PlayURL` (`Ahu_Course.pyc.1shot.das:11130-11367`) | `aliyunPlayInfoResp.PlayInfoList.PlayInfo[].PlayURL` (`ahu.go:181-188`) | ✓ |
| `PlayInfoList -> PlayInfo -> Definition` | `aliyunPlayInfoResp.PlayInfoList.PlayInfo[].Definition` (`ahu.go:183-186`) | ✓ |
| `AccessKeyId / AccessKeySecret / SecurityToken / Region / AuthInfo / AuthTimeout` (`Ahu_Course.pyc.1shot.das:10186-10671`) | `aliyunPlayAuth` tags (`ahu.go:165-179`) | ✓ |

## 阻塞步骤

无。
