# 项目：个人运动打卡API

从0做一个Go个人运动打卡API，用Gin开发，SQLite存储。记录包含运动类型、运动时长、消耗热量、日期和备注，支持新增修改删除、按日期范围查询、按运动类型汇总时长和热量、查看最近一周打卡情况。代码遵循Go企业分层结构：cmd/server/main.go、internal/config、internal/model、internal/repository、internal/service、internal/handler、internal/router、internal/middleware、migrations。热量计算规则放在service层，查询接口支持分页。
