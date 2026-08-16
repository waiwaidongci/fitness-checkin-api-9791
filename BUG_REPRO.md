# Bug 是什么

单条打卡的 summary 接口从 `Workout.Summary` 读取汇总对象，但查询数据库时没有填充该指针，接口处理时直接访问其字段，触发 nil 指针解引用。

# 如何触发

```bash
go test ./internal/handler -run TestWorkoutSummaryEndpoint -count=20
```

# 错误信息

```text
runtime error: invalid memory address or nil pointer dereference
(*WorkoutHandler).WorkoutSummary
```
