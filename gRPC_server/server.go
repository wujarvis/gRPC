package main

import (
	"context"
	"fmt"
	"gRPC/proto/hi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"log"
	"net"
)

// gRPC 服务地址
const Address = "127.0.0.1:8080"

// 定义Hi结构体并实现约定接口
type Hi struct {
}

// SayHi
func (h *Hi) SayHi(ctx context.Context, request *hi.HiRequest) (*hi.HiResponse, error) {
	resp := new(hi.HiResponse)
	resp.Message = fmt.Sprintf("Hi %s", request.Name)

	return resp, nil
}

// 认证token
func myAuth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return grpc.Errorf(codes.Unauthenticated, "no token ")
	}

	log.Println("myAuth ...")

	var (
		appId  string
		appKey string
	)

	// md 是一个 map[string][]string 类型的
	if val, ok := md["appid"]; ok {
		appId = val[0]
	}

	if val, ok := md["appkey"]; ok {
		appKey = val[0]
	}

	if appId != "myappId" || appKey != "myKey" {
		return grpc.Errorf(codes.Unauthenticated, "token invalide: appid=%s, appkey=%s", appId, appKey)
	}

	return nil
}

// interceptor 拦截器
/*
// If a UnaryHandler returns an error, it should either be produced by the
// status package, or be one of the context errors. Otherwise, gRPC will use
// codes.Unknown as the status code and err.Error() as the status message of the
// RPC.
type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

// UnaryServerInterceptor provides a hook to intercept the execution of a unary RPC on the server. info
// contains all the information of this RPC the interceptor can operate on. And handler is the wrapper
// of the service method implementation. It is the responsibility of the interceptor to invoke handler
// to complete the RPC.
type UnaryServerInterceptor func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error)
*/
func interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 认证
	log.Println("interceptor...")
	err := myAuth(ctx)
	if err != nil {
		return nil, err
	}

	// 继续处理请求
	return handler(ctx, req)
}

func main() {
	log.SetFlags(log.Ltime | log.Llongfile)

	listen, err := net.Listen("tcp", Address)
	if err != nil {
		log.Panicf("faild to listen: %v", err)
	}

	//var opts []grpc.ServerOption

	opts := make([]grpc.ServerOption, 0)

	// TLS 认证
	creds, err := credentials.NewServerTLSFromFile("../keys/server.pem", "../keys/server.key")
	if err != nil {
		log.Panicf("faild to generate credentials %v", err)
	}

	// 实例化gprc server, 开启TLS认证
	opts = append(opts, grpc.Creds(creds))

	// 注册拦截器
	opts = append(opts, grpc.UnaryInterceptor(interceptor))

	// 实例化grpc Server, 并开启TLS认证，其中还有拦截器
	s := grpc.NewServer(opts...)

	// 注册HelloService
	hi.RegisterHiServer(s, new(Hi))

	log.Println("Listen on" + Address + "with TLS")

	s.Serve(listen)
}
