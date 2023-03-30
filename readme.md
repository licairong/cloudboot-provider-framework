# cloudboot-provider-framework
CloudBoot provider基础框架

使用方法
在plugin/plugin.go中引入相关module，执行编译：
```shell
go build -o dell-poweredge-r730 plugin/plugin.go
```

运行server/server.go

# proto generate
```shell
protoc -I proto/ proto/print.proto --go_out=plugins=grpc:proto/ --go_out=. --go_opt=paths=source_relative

# 分离grpc 的代码到单独文件
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/plugin.proto
```