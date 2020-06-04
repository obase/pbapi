# package pbapi
包含api协定框架的元数据. 用户实现service, 轻松提供http, websocket, grpc等多种访问渠道. 


## 支持优雅关闭/重启:
1. graceful shutdown: windows, linux, darwin
```
kill -HUP/-INT/-TERM <pid>, 或者kill <pid>
```
2. graceful restart: linux, darwing
```
kill -USR2 <pid>
```

## api框架的目录结构:
```
$project
   |__bin
   |__pkg
   |__src
   |  |__api: 接口proto及自动生成文件,并使用pbapigen维护基础代码.
   |  |__biz: 业务实现逻辑.
   |  |__dao: 数据持久逻辑.
   |  |__mdl: 模型数据结构.
   |  |__main.go: 用户注册service实例
   |  |__conf.yml.template: 服务配置模板
   |
   |__conf.yml: 本地测试配置
      
最基本的项目结构, 另根据需求添加dao, model等业务package
```

## pbapi框架的辅助工具pbapigen
pbapigen用于api框架自动代码生成工具, 能大大减少代码编写量!

- windows(64位)版本: https://obase.github.io/pbapigen/windows/pbapigen.exe
- linux(64位)版本:   https://obase.github.io/pbapigen/linux/pbapigen
- darwin(64位)版本:  https://obase.github.io/pbapigen/darwin/pbapigen

其他平台请从源码编译:
- go get -u github.com/obase/pbapigen
- go install github.com/obase/pbapigen

在$GOPATH/bin查找编译后可执行文件pbapigen

注意: pbapigen执行过程会生成".pbapigen"目录, 自动下载需要用到的protoc工具与protoc-gen-pbapi工具. 目录结构类似:
```
<PATH>
  |__pbapigen
  |__.pbapigen/
     |__protoc.exe
     |__protc-gen-pbapi.exe
     |__version
```

## pbapi框架的使用步骤
1. 创建项目目录结构(见上), 所有接口proto文件必须放在api及其子包里.
2. 打开DOS或Shell, 进入src目录, 即api父目录.
```
cd $project/src/apidemo
```
3. 执行apigen命令,自动生成proto的代码文件并存于api对应目录内.
```
pbapigen
```
如果不在api所在的目录,则需要用"pbapigen -parent=xxx"指定, 例如:
```
C:\>pbapigen -parent C:\Workspace\pbapiworkspace\src\github.com\obase\apidemo
```
注意: 请不要手工修改$project/src/api目录里面的*.pb.go文件内容, 避免操作被pbapigen结果覆盖.

## pbapi框架的局限地方

基于性能考虑, apix使用标准encoding/json(而非grpc的jsonpb)处理protobuf的json. 
经测试不支持protobuf的Onceof, Any等高级特性(除非使用jsonpb)!

# Installation
- go get
```
go get -u github.com/gin-gonic/gin
go get -u github.com/golang/protobuf
go get -u github.com/gorilla/websocket
go get -u google.golang.org/grpc 
go get -u github.com/obase/pbapi
go get -u github.com/obase/center 
go get -u github.com/obase/conf 
go get -u github.com/obase/log
```
- go mod
```
go mod edit -require=github.com/obase/pbapi@latest
```
强烈建议go mod, 自动级联下载所需依赖

# Configuration
```
# 服务元数据
service:
  # 服务名称, 自动注册<name>, <name>.http, <name>.grpc三种服务
  name: "demo"
  # Http请求(post请求及websocket请求)主机, 如果为空, 默认本机首个私有IP
  httpHost: "127.0.0.1"
  # Http请求(post请求及websocket请求)端口, 如果为空, 则不启动Http服务器
  httpPort: 8000
  # consul健康检查超时及间隔. 默认5s与6s
  httpCheckTimeout: "5s"
  httpCheckInterval: "6s"
  # Grpc请求主机, 如果为空, 默认本机首个私有IP
  grpcHost: "127.0.0.1"
  # Grpc请求端口, 如果为空, 则不启动Grpc服务器
  grpcPort: 8100
  # consul健康检查超时及间隔
  grpcCheckTimeout: "5s"
  grpcCheckInterval: "6s"
  # 启动模式: DEBUG, TEST, RELEASE
  mode: "DEBUG"
  # Weboscket读缓存大小
  wsReadBufferSize: 8092
  # Websocket写缓存大小
  wsWriteBufferSize: 8092
  # Websocket不校验origin
  wsNotCheckOrigin: false
  # consult 配置地址, 默认不启动.也支持
  # center:
  #   address: "127.0.0.1:8500"
  #   timeout: "30s"
```

- func (server *Server) Serve() 
```
func (server *Server) Serve() error 
```
启动服务

# Examples
proto
```
syntax = "proto3";

package api;

import "github.com/obase/api/x.proto";

// grpc的白名单过滤器
option (server_option) = {pack:"github.com/obase/demo/system" func:"AccessGuarderGrpc"};
// http的Logger过滤器
option (middle_filter) = {pack:"github.com/obase/demo/system" func:"AccessLoggerHttp"};

service IPlayer {
    // post分组, 配套还有group_filter
    option (group) = {path:"/player"};

    rpc Add (Player) returns (Player) {
        // post请求, 因为配置了group path, 所以路径为/player/add
        option (handle) = {path:"/add"};
        // websocket请求, 因为配置了group path, 所以路径为/player/add
        option (socket) = {path:"/add"};
    }
    rpc Del (Player) returns (void){
        // post请求, 因为配置了group path, 所以路径为/player/del
        option (handle) = {path:"/del"};
        // websocket请求, 因为配置了group path, 所以路径为/player/del
        option (socket) = {path:"/del"};
    }
    rpc Get (Player) returns (Player){
        // post请求, 因为配置了group path, 所以路径为/player/get
        option (handle) = {path:"/get"};
        // websocket请求, 因为配置了group path, 所以路径为/player/get
        option (socket) = {path:"/get"};
    }
    rpc List (void) returns (PlayerList){
        // post请求, 因为配置了group path, 所以路径为/player/list
        option (handle) = {path:"/list"};
        // websocket请求, 因为配置了group path, 所以路径为/player/list
        option (socket) = {path:"/list"};
    }
}

// 默认字段
message Player {
    string id = 1;
    string name = 2;
    string globalRoleId = 3;
}

message PlayerList {
    repeated Player players = 1;
}

service ICorps {
    // post分组, 配套还有group_filter
    option (group) = {path:"/corps"};

    rpc Add (Corps) returns (void) {
        // post请求, 因为配置了group path, 所以路径为/corps/add
        option (handle) = {path:"/add"};
        // websocket请求, 因为配置了group path, 所以路径为/corps/add
        option (socket) = {path:"/add"};
    }
    rpc Del (Corps) returns (void){
        // post请求, 因为配置了group path, 所以路径为/corps/del
        option (handle) = {path:"/del"};
        // websocket请求, 因为配置了group path, 所以路径为/corps/del
        option (socket) = {path:"/del"};
    }
    rpc Get (Corps) returns (Corps){
        // post请求, 因为配置了group path, 所以路径为/corps/get
        option (handle) = {path:"/get"};
        // websocket请求, 因为配置了group path, 所以路径为/corps/get
        option (socket) = {path:"/get"};
    }
    rpc List (void) returns (CorpsList){
        // post请求, 因为配置了group path, 所以路径为/corps/list
        option (handle) = {path:"/list"};
        // websocket请求, 因为配置了group path, 所以路径为/corps/list
        option (socket) = {path:"/list"};
    }
}

message Corps {
    string id = 1;
    string name = 2;
    string logo = 3;
    fixed32 type = 4; // 2, 3, 5
}

message CorpsList {
    repeated Corps corps = 1;
}

```

codes:
```
func main() {
	server := apix.NewServer()
	// 注册服务
	api.RegisterICorpsService(server, &service.ICorpsService{})
	api.RegisterIPlayerService(server, &service.IPlayerService{})

	// 启动服务
	if err := server.ServeWith(&apix.Conf{
                               		HttpHost:         "127.0.0.1",
                               		HttpPort:         8000,
                               		GrpcHost:         "127.0.0.1",
                               		GrpcPort:         9000,
                               		WsNotCheckOrigin: true, //不检查websocket的Origin,方便测试
                               	}); err != nil {
		panic(err)
	}
}

```