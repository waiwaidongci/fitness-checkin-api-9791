# Bug 是什么

最近一周打卡接口新增了内存缓存，后台刷新 goroutine 和请求 goroutine 都会更新同一个 `[]model.Workout` 切片，但缓存结构没有同步保护，`Snapshot` 也直接返回内部切片，存在并发读写/写入竞争。

# 如何触发

```bash
go test -race ./internal/service -run TestRecentWeekConcurrent -count=20
```

# 错误信息

```text
WARNING: DATA RACE
Write at 0x00c0000101e0 by goroutine ...
  internal/cache.(*RecentWeek).Replace()
Previous write at 0x00c0000101e0 by goroutine ...
  internal/cache.(*RecentWeek).Replace()
--- FAIL: TestRecentWeekConcurrent
```
