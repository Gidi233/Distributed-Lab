# 简易分布式资源管理系统
分布式的大作业具体要求在ppt。
可能在 `isElection` 那儿还有点问题。

## 构建
```bash
kitex -module main -service main ops.thrift
go build main.go
./main <port>
```