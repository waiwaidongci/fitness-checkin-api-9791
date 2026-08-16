# Bug 是什么

重复打卡被业务层识别后返回了普通错误而不是稳定的 sentinel 错误，HTTP 错误映射无法识别该错误，最终返回 500 而不是 409 Conflict。

# 如何触发

```bash
go test ./internal/handler -run TestDuplicateWorkoutReturnsConflict -count=20
```

# 错误信息

```text
--- FAIL: TestDuplicateWorkoutReturnsConflict
    expected 409 conflict, got 500
```
