# Bug 是什么

列表查询的 context 在 handler、service、repository 三层都被丢弃：handler 和 service 逐层改成 `context.Background()`，repository 又使用普通 `Query` / `QueryRow`，导致已取消的 context 不会让 SQLite 查询失败。

# 如何触发

```bash
go test ./...
```

# 错误信息

```text
--- FAIL: TestListContextHonorsCancellation
    expected list to honor the canceled context
```
