---
title: Golang使用grpc指南
date: 2020-09-04
tags: [protobuf, grpc, grpc-gateway, RestApi]
categories: [编程, Golang, libs&framworks]
type: DEMO

---
# Golang使用 grpc 指南
[toc]

本文以一个简单的CURD服务为例演示了如果一步步使用grpc的接口. 

## 使用protobuf

### 编写proto文件
proto文件是定义整个protobuf生态的基石, protobuf,grpc, grpc-gateway等代码都是通过proto文件来生成桩代码的.  
proto文件主要包含 syntax, package, option, import, message, service 等几部分.  
- syntax  
  用于指定proto文件使用了protobuf协议的版本. 可以选择proto2或proto3, 推荐使用较新的版本proto3;
- package  
  用于指定当前proto文件所在的包名称, 只需要声明完整包名的最后一部分即可. 例如某个文件`A.proto`位于`github.com/someproject/api/Apple.proto`, 那么这个文件的package只需要指定some_apis即可.  
- option  
  用于针对特定场景设定一些选项. 例如在v1.20以后的版本中, 用户如果想要将proto编译成Golang代码, 就需要指定`go_package`选项. 
- import  
  proto文件可以分成多个, 不同的proto文件之间可以使用import字段互相引用, 达到代码复用的目的. import声明的路径可以是相对路径, 例如如果引入同一个目录下的其他proto文件, 则可以直接写文件名. 如果引入其他目录的文件, 也不必从文件系统根目录或项目根目录开始写起, 但是在编译的时候必须通过`-I`参数指明找到被引入文件的起始路径, 换句话说 `-I`+`import` 需要指向proto文件的路径. 
- message  
  message用于定义数据结构.  如果不使用grpc远程调用, 而只是用protobuf作为数据传输的格式的话,  那么只需要定义message即可, 不需要定义service. 
- service  
  用于定义API服务的通信接口. 

我们先只用protobuf, 因此只需要不需要声明service, 只需要定义message就够了. 
```proto
syntax = "proto3";
package api;
option go_package="github.com/violin0622/grpc-apple/api";

message Apple{
  int32 number = 1;
  string name = 2;
  Size size = 3;

  enum Size{
    SIZE_UNDEFINED = 0;
    BIG = 4;
    MID = 5;
    SMALL = 6;
  }
}
```

### 下载protoc工具
protoc 是用于编译proto文件生成对应桩代码的命令行工具. protoc工具使用C++编写, 其项目地址位于[https://github.com/protocolbuffers/protobuf ](https://github.com/protocolbuffers/protobuf) .   
对于非C++的用户, 可以直接下载预先编译好的二进制文件: [https://github.com/protocolbuffers/protobuf/releases ](https://github.com/protocolbuffers/protobuf/releases), 例如对于macOS用户, 选择 `protobuf-3.13.0-osx-x86_64.tar.gz`下载解压, 并放在$PATH下面. 

### 下载protoc插件: protoc-gen-go
protoc编译proto文件时, 根据生成不同语言的桩代码的需求, 需要指定不同的插件. 例如需要生成Golang的桩代码, 便需要指定go语言的插件: `protoc-gen-go`.  
值得一提的是, golang的插件在2020年初经历了比较大的变更: 原来其项目地址位于[github.com/golang/protobuf](https://github.com/golang/protobuf), 代码中的导入地址是 `github.com/golang/protobuf/proto`, 其版本迭代到 v1.4,  2020年三月份由新的项目取代: [google.golang.org/protobuf](https://github.com/protocolbuffers/protobuf-go), 版本从 v1.20 开始迭代. v1.20相对于v1.4作出了许多重大的变更, 包括部分API变更, 以及原有的部分可导出模块不再导出. 对于中国开发者来说比较重要的变更在于修复了打印结构体内的非ASCII字符会乱码的bug( [Issue #572](https://github.com/golang/protobuf/issues/572) ).  
插件可以直接在项目的Github Release页面下载编译好的版本, 或者 git clone 然后自行安装:
```sh
git clone -b v1.31 git@github.com:protocolbuffers/protobuf-go.git
cd protobuf-go
# go install 会将项目编译生产的二进制文件放入 $GOPATH/bin. 
# 需要把 $GOPATH/bin 加入 $PATH 以使protoc能够找到. 
go install .
```

### 编译
编译使用的命令参数可以见另一篇文章.
```sh
protoc \
  --go_out=paths=source_relative:. \
  api/apple.proto
```
可以在api目录下看到生成的桩代码 `apple.pb.go`
```
├── README.md
├── api
│   ├── apple.pb.go
│   └── apple.proto
├── go.mod
└── main.go
```

### 使用
为了使用 apple.pb.go, 我们还需要在Golang代码中导入protobuf的运行库 `google.golang.org/protobuf/proto`, 然后利用其中的`Marshal`和`Unmashal` 两个API 进行消息的编码解码. 
[examples](https://github.com/protocolbuffers/protobuf/blob/master/examples/add_person.go)
```go
package main
import(
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

  `google.golang.org/protobuf/proto`

  `github.com/grpc-apple/api`
)

func main(){
	if len(os.Args) != 2 {
		log.Fatalf("Usage:  %s ADDRESS_BOOK_FILE\n", os.Args[0])
	}
	fname := os.Args[1]

	// Read the existing address book.
	in, err := ioutil.ReadFile(fname)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s: File not found.  Creating new file.\n", fname)
		} else {
			log.Fatalln("Error reading file:", err)
		}
	}

	// [START marshal_proto]
	book := &pb.AddressBook{}
	// [START_EXCLUDE]
	if err := proto.Unmarshal(in, book); err != nil {
		log.Fatalln("Failed to parse address book:", err)
	}

	// Add an address.
	addr, err := promptForAddress(os.Stdin)
	if err != nil {
		log.Fatalln("Error with address:", err)
	}
	book.People = append(book.People, addr)
	// [END_EXCLUDE]

	// Write the new address book back to disk.
	out, err := proto.Marshal(book)
	if err != nil {
		log.Fatalln("Failed to encode address book:", err)
	}
	if err := ioutil.WriteFile(fname, out, 0644); err != nil {
		log.Fatalln("Failed to write address book:", err)
	}
	// [END marshal_proto]
}
```

## 使用grpc
我们需要查询, 创建, 更新, 修改, 删除五个接口. 
### 编写proto文件
事实上, 所有的定义都可以写进一个proto文件里, 不过在复杂项目中这样显然不好. 本文使用了多个不同目录下的proto文件, 以演示在复杂项目中不同的定义是怎样互相引用的.  
此处我们为了定义五个接口而在新的目录`api/operation/`创建了新的`operations.proto`文件.  
值得一提的是导入`apple.proto`文件时指定的路径. 上文已经说了import与-I参数的关系, 实际上这两个参数还关系着生成的桩代码的位置. 在本项目中, 希望将桩代码与对应的proto文件放在一起.  
```proto
syntax = "proto3";
package operation;
option go_package="github.com/violin0622/grpc-apple/api/operation";
// 引入protobuf官方提供的Empty结构作为部分接口的空返回值. 
import "google/protobuf/empty.proto";
// 引入protobuf官方提供的Field_Mask用于支持修改操作. 
import "google/protobuf/field_mask.proto";
// 引入上级目录的apple文件以复用定义
import "api/apple.proto";

service AppleService{
  rpc DescribeApple(DescribeAppleRequest) returns (api.Apple) {}
  rpc CreateApple(CreateAppleRequest) returns (api.Apple) {}
  rpc UpdateApple(UpdateAppleRequest) returns (api.Apple) {}
  rpc ModifyApple(ModifyAppleRequest) returns (api.Apple) {}
  rpc DestroyApple(DestroyAppleRequest) returns (google.protobuf.Empty) {}
}

// 为了提供调用接口, 我们新声明了五个消息类型, 需要定义. 
message DescribeAppleRequest{
  int32 number = 1;
}
message CreateAppleRequest{
  string name = 2;
  api.Apple.Size size = 3;
}
// 更新操作: 必须指定对象全部的属性, 
// 对于未指定的属性, 应该将其设定为空或默认值; 
message UpdateAppleRequest{
  int32 number = 1;
  string name = 2;
  api.Apple.Size size = 3;
}
// 修改操作: 只需要设定对象需要变更的属性, 
// 对于未指定的属性, 会保留原来的值. 
// grpc中为了支持修改操作, 需要添加额外的FieldMask字段. 
// 不过好在该字段的值不需要用户设定, grpc会自动生成. 
message ModifyAppleRequest{
  int32 number = 1;
  string name = 2;
  api.Apple.Size size = 3;
  google.protobuf.FieldMask mask = 4;
}

message DestroyAppleRequest{
  string name = 1;
}
```

### 下载protoc插件: protoc-gen-go-grpc
同 protoc-gen-go 一样, 生成grpc代码也需要对应的插件: protoc-gen-go-grpc. 在`github.com/golang/protobuf`将protoc-gen-go代码的归属权托管给了`golang.google.org/grpc-go`(参考[golang/protobuf #903](https://github.com/golang/protobuf/issues/903)), 从此作为grpc-go的一个工具, 生成grpc桩代码的方式也发生了巨大的变化.  
在protoc-go v1.4 版本之前, 也就是旧项目中, protoc-grpc是作为 protoc-gen-go的插件存在的, 也就是protoc的插件的插件.  
而在 protoc-go v1.20 之后, 也就是新项目中, protoc-grpc是作为protoc的插件存在, 也就是“升级”了, 从插件的插件变成了独立的插件.  
proto-gen-go-grpc 目前还没有发布预编译的新插件, 想要使用的话必须自行编译安装:
```sh
# 直接安装:
go get -u google.golang.org/grpc

# 或者这样:
git clone git@github.com:grpc/grpc-go.git
cd grpc-go && go install .
```

### 编译
由于存在新旧两种插件, 因此编译命令也有了两种.  
值得一提的是, protoc 不支持一次性编译多个包, 如果指定了多个包, 会造成错误.  
旧版命令:  
```sh
protoc \
    --go_out=plugins=grpc,paths=source_relative:. \
    api/apple.proto 

protoc \
  -Iapi
  --go_out=plugins=grpc,paths=source_relative:. \
  api/operation/operations.proto
```
新版命令:
```sh
protoc \
  --go_out=paths=source_relative:. \
  --go-grpc_out=paths=source_relative:. \
  api/apple.proto

protoc \
  --go_out=paths=source_relative:. \
  --go-grpc_out=paths=source_relative:. \
  api/operation/operations.proto
```
显然新版命令比旧版的更长了 -,-!

编译之后可以看到proto文件所在目录下多了生成的桩代码. 
`apple.pb.go`是为api/apple.proto生成的; `operations.pb.go` 是为operations.proto中的message生成的; `operations_grpc.pb.go` 是为operations.proto中service的部分生成的. 为什么message和service要分别生成两个文件, 我也不知道.  
```
├── README.md
├── api
│   ├── apple.pb.go
│   ├── apple.proto
│   └── operation
│       ├── operations.pb.go
│       ├── operations.proto
│       └── operations_grpc.pb.go
├── go.mod
└── main.go
```

### 实现接口
新建一个service 包, 将几个接口的具体实现放在里面:  
service/service.go:  
```go
package service

import (
	"context"
	"log"

	"github.com/golang/protobuf/ptypes/empty"

	. "github.com/violin0622/grpc-apple/api/operation"
	. "github.com/violin0622/grpc-apple/api"
)

// protoc-gen-go-grpc v1.20 之后, 实现服务的时候必须嵌入 UnimplementedAppleServiceServer 以保证能够向后兼容.
type AppleService struct {
	UnimplementedAppleServiceServer
}

func (*AppleService) DescribeApple(ctx context.Context, req *DescribeAppleRequest) (*Apple, error) {
	return &Apple{}, nil
}
func (*AppleService) CreateApple(ctx context.Context, req *CreateAppleRequest) (*Apple, error) {
	log.Println(req)
	return &Apple{}, nil
}
func (*AppleService) UpdateApple(ctx context.Context, req *UpdateAppleRequest) (*Apple, error) {
	return &Apple{}, nil
}
func (*AppleService) ModifyApple(ctx context.Context, req *ModifyAppleRequest) (*Apple, error) {
	return &Apple{}, nil
}
func (*AppleService) DestroyApple(ctx context.Context, req *DestroyAppleRequest) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}
```

### 使用
代码位于`simple-grpc`:  
```go
package main

import (
	"context"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/violin0622/grpc-apple/api"
	. "github.com/violin0622/grpc-apple/api/operation"
)


func main() {
	grpcServer := grpc.NewServer()
	RegisterAppleServiceServer(grpcServer, &AppleService{})
	reflection.Register(grpcServer)

	if l, err := net.Listen(`tcp`, `:9000`); err != nil {
		log.Fatal(`cannot listen to port 9000: `, err)
	} else if err = grpcServer.Serve(l); err != nil {
		log.Fatal(`cannot start service:`, err)
	}
}
```

## 使用grpc-gateway
[grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) 是grpc-ecosystem 的子项目, 用于提供grpc的反向代理, 实现开发GRPC, 提供RestAPI的目的.  
grpc-gateway 同样也是 protoc 的插件, 并且仅支持生成Golang语言的桩代码.  
但是, 生成的反向代理可以作为独立的进程, 因此实际可以支持各种语言的grpc服务.  只不过作为Golang语言可以实现更多的特性, 比如复用端口同时提供grpc和http两种接口. 

### 修改operations.proto
为了使用grpc-gateway, 需要在 operations.proto 中引入grpc-gateway的proto文件, 并在每个rpc中添加配置.  
```proto
syntax = "proto3";
package operation;
option go_package="github.com/violin0622/grpc-apple/operation/api";
// 引入protobuf官方提供的Empty结构作为部分接口的空返回值. 
import "google/protobuf/empty.proto";
// 引入protobuf官方提供的Field_Mask用于支持修改操作. 
import "google/protobuf/field_mask.proto";
// 引入annotation用于定义gateway
import "google/api/annotations.proto";
// 引入上级目录的apple文件以复用定义
import "apple.proto";

service AppleService{
  rpc DescribeApple(DescribeAppleRequest) returns (Apple) {
    option (google.api.http).get = "/apples/{number}";
  }
  rpc CreateApple(CreateAppleRequest) returns (Apple) {
    option (google.api.http) = {
      post: "/apples"
      body: "*"
    };
  }
  rpc UpdateApple(UpdateAppleRequest) returns (Apple) {
    option (google.api.http) = {
      put: "/apples/{number}"
      body: "*"
    };
  }
  rpc ModifyApple(ModifyAppleRequest) returns (Apple) {
    option (google.api.http) = {
      patch: "/apples/{number}"
      body: "*"
    };
  }
  rpc DestroyApple(DestroyAppleRequest) returns (google.protobuf.Empty) {
    option (google.api.http).delete = "/apples/{number}";
  }
}

// 为了提供调用接口, 我们新声明了五个消息类型, 需要定义. 
message DescribeAppleRequest{
  int32 number = 1;
}
message CreateAppleRequest{
  string name = 2;
  Apple.Size size = 3;
}
// 更新操作: 必须指定对象全部的属性, 
// 对于未指定的属性, 应该将其设定为空或默认值; 
message UpdateAppleRequest{
  int32 number = 1;
  string name = 2;
  Apple.Size size = 3;
}
// 修改操作: 只需要设定对象需要变更的属性, 
// 对于未指定的属性, 会保留原来的值. 
// grpc中为了支持修改操作, 需要添加额外的FieldMask字段. 
// 不过好在该字段的值不需要用户设定, grpc会自动生成. 
message ModifyAppleRequest{
  int32 number = 1;
  string name = 2;
  Apple.Size size = 3;
  google.protobuf.FieldMask mask = 4;
}

message DestroyAppleRequest{
  int32 number = 1;
}
```

### 下载 protoc-gen-grpc-gateway
```sh
go get github.com/grpc-ecosystem/grpc-gateway
```

### 编译
apple.proto 我们已经编译过了并且没有修改过, 因此可以直接编译 operations.proto
```sh
protoc \
    -I. \
    -I$GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.14.8/third_party/googleapis  \
    --go_out=paths=source_relative:. \
    --go-grpc_out=paths=source_relative:. \
    --grpc-gateway_out=paths=source_relative:. \
    api/operation/*.proto
```

### 使用
由于grpcServer监听端口会把线程Hang住, 因此需要通过`go`操作符创建额外的goroutine用于运行监听函数.  
程序运行起来之后会监听两个端口: grpc服务监听本地8000端口, http服务监听本地9000端口, 并把请求转发到8000端口的grpc服务上.  
此时可以使用 curl localhost:8000 通过http方式访问, 也可以使用 grpcurl -plaintext localhost:9000 通过grpc方式访问.  
```go
package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	. "github.com/violin0622/grpc-apple/api/operation"
	. "github.com/violin0622/grpc-apple/api"
)

// protoc-gen-go-grpc v1.20 之后, 实现服务的时候必须嵌入 UnimplementedAppleServiceServer 以保证能够向后兼容.
type AppleService struct {
	UnimplementedAppleServiceServer
}

func (*AppleService) DescribeApple(ctx context.Context, req *DescribeAppleRequest) (*Apple, error) {
	return &Apple{}, nil
}
func (*AppleService) CreateApple(ctx context.Context, req *CreateAppleRequest) (*Apple, error) {
	log.Println(req)
	return &Apple{}, nil
}
func (*AppleService) UpdateApple(ctx context.Context, req *UpdateAppleRequest) (*Apple, error) {
	return &Apple{}, nil
}
func (*AppleService) ModifyApple(ctx context.Context, req *ModifyAppleRequest) (*Apple, error) {
	return &Apple{}, nil
}
func (*AppleService) DestroyApple(ctx context.Context, req *DestroyAppleRequest) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func main() {
	grpcServer := grpc.NewServer()
	RegisterAppleServiceServer(grpcServer, &AppleService{})
	reflection.Register(grpcServer)
	l, _ := net.Listen(`tcp`, `:8000`)
	go grpcServer.Serve(l)

	httpServer := runtime.NewServeMux()
	RegisterAppleServiceHandlerFromEndpoint(
		context.Background(),
		httpServer,
		`:8000`,
		[]grpc.DialOption{grpc.WithInsecure()},
	)

	if err := http.ListenAndServe(`:9000`, httpServer); err != nil {
		log.Fatal(`cannot start service: `, err)
	}
}

```

## gateway 与 grpc 使用相同的端口(使用TLS)
基本思路是通过判断 Content-Type 字段来分辨入请求是基于HTTP还是GRPC, 然后分别转发到对应的server handler上.  
代码位于 reuse-port-tls/server.go
```go
package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	. "github.com/violin0622/grpc-apple/api/operation"
	"github.com/violin0622/grpc-apple/service"
)

func main() {
	serverCred, err := credentials.NewServerTLSFromFile(`./server.pem`, `./server.key`)
	if err != nil {
		log.Fatal(err)
	}
	clientCred, err := credentials.NewClientTLSFromFile(`./server.pem`, `localhost`)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(serverCred))
	RegisterAppleServiceServer(grpcServer, &service.AppleService{})
	reflection.Register(grpcServer)

	httpServer := runtime.NewServeMux()
	RegisterAppleServiceHandlerFromEndpoint(
		context.Background(),
		httpServer,
		`:8000`,
		[]grpc.DialOption{grpc.WithTransportCredentials(clientCred)},
	)

	http.ListenAndServeTLS(`:8000`, `./server.pem`, `./server.key`,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.ProtoMajor == 2 &&
				strings.Contains(r.Header.Get(`Content-Type`), `application/grpc`) {
				log.Println(`grpc`)
				grpcServer.ServeHTTP(w, r)
			} else {
				log.Println(`http`)
				httpServer.ServeHTTP(w, r)
			}
		}),
	)
}
```

## gateway 与 grpc 使用相同的端口(不使用TLS)
代码位于 reuse-port-insecure/server.go.  
```go
package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	. "github.com/violin0622/grpc-apple/api/operation"
	"github.com/violin0622/grpc-apple/service"
)

func main() {
	grpcServer := grpc.NewServer()
	RegisterAppleServiceServer(grpcServer, &service.AppleService{})
	reflection.Register(grpcServer)

	httpServer := runtime.NewServeMux()
	RegisterAppleServiceHandlerFromEndpoint(
		context.Background(),
		httpServer,
		`:8000`,
		[]grpc.DialOption{grpc.WithInsecure()},
	)

	http.ListenAndServe(
		`:8000`,
		h2c.NewHandler(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				if r.ProtoMajor == 2 &&
					strings.Contains(r.Header.Get(`Content-Type`), `application/grpc`) {
					log.Println(`grpc`)
					grpcServer.ServeHTTP(w, r)
				} else {
					log.Println(`http`)
					httpServer.ServeHTTP(w, r)
				}
			}),
			&http2.Server{}),
	)
}
```
