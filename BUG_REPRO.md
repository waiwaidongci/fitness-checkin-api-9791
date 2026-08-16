# Bug 是什么

SQL 迁移执行器用 `strings.Split(sql, ";")` 切分语句，遇到字符串字面量或 CHECK 约束中的分号会把一条 SQL 拆坏，导致迁移执行失败。

# 如何触发

```bash
go test ./...
```

# 错误信息

```text
open repository: run migration 002_sql_text.sql: SQL logic error: unrecognized token: "'" (1)
```
