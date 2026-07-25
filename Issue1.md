# FWAlizer 构建问题记录

> 记录构建过程中遇到的错误、排查过程及解决办法。

---

## Issue 1: go.mod 重复内容

**Step:** 1

**现象：** `go build ./...` 报错 `repeated module statement` 和 `repeated go statement`

**原因：** 覆写 go.mod 时旧内容未完全清除，导致 module 和 go 声明重复出现

**解决：** 重新完整写入 go.mod，确保仅包含一份 module 声明和 go 版本声明

---

## Issue 2: RULES 端口字段逗号与条目分隔符冲突

**Step:** 2

**现象：** `RULES=api.example.com|TCP|443,80|ACCEPT||生产API` 解析失败，端口 `443,80` 中的逗号被误认为条目分隔符

**原因：** Build1.md 参考代码使用简单逗号拆分 (`splitEntries`)，但端口字段允许逗号分隔多端口

**解决：** 实现 `splitRuleEntries` 智能分割——通过正则检测 `host|PROTOCOL|` 模式识别新条目起始位置，仅在该模式匹配时才分割，否则将逗号保留在当前条目内（属于端口字段）
