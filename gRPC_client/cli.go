package main

import (
	pb "gRPC/proto/hi" // 引入proto包
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials" // 引入grpc认证包
	"google.golang.org/grpc/grpclog"
)

const (
	// Address gRPC服务地址
	Address = "127.0.0.1:8080"
)

// ==== start ====
var IsTLS = true

// 自定义认证
type myCredential struct {
}

/*
/ PerRPCCredentials defines the common interface for the credentials which need to
// attach security information to every RPC (e.g., oauth2).
type PerRPCCredentials interface {
	// GetRequestMetadata gets the current request metadata, refreshing tokens
	// if required. This should be called by the transport layer on each
	// request, and the data should be populated in headers or other
	// context. If a status code is returned, it will be used as the status for
	// the RPC (restricted to an allowable set of codes as defined by gRFC
	// A54). uri is the URI of the entry point for the request.  When supported
	// by the underlying implementation, ctx can be used for timeout and
	// cancellation. Additionally, RequestInfo data will be available via ctx
	// to this call.  TODO(zhaoq): Define the set of the qualified keys instead
	// of leaving it as an arbitrary string.
	GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error)
	// RequireTransportSecurity indicates whether the credentials requires
	// transport security.
	RequireTransportSecurity() bool
}
*/
// 实现自定义认证接口
func (c *myCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"appId":  "myappId",
		"appKey": "myKey",
	}, nil
}

// 自定义认证是否开启tls
func (c *myCredential) RequireTransportSecurity() bool {
	return IsTLS
}

// 客户端拦截器
func clientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Printf("method == %s ; req == %v ; rep == %v ; duration == %s ; error == %v\n", method, req, reply, time.Since(start), err)
	return err
}

func main() {
	log.SetFlags(log.Ltime | log.Llongfile)

	var err error
	var opts []grpc.DialOption
	// TLS连接  记得把xxx改成你写的服务器地址
	if IsTLS {
		creds, err := credentials.NewClientTLSFromFile("../keys/server.pem", "www.eline.com")
		if err != nil {
			log.Panicf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// 自定义认证，new(myCredential 的时候，由于我们实现了上述2个接口，因此new的时候，程序会执行我们实现的接口
	opts = append(opts, grpc.WithPerRPCCredentials(new(myCredential)))

	// 加上拦截器
	opts = append(opts, grpc.WithUnaryInterceptor(clientInterceptor))

	conn, err := grpc.Dial(Address, opts...)
	if err != nil {
		grpclog.Fatalln(err)
	}

	defer conn.Close()

	// 初始化客户端
	c := pb.NewHiClient(conn)

	// 调用方法
	req := &pb.HiRequest{Name: "gRPC"}
	res, err := c.SayHi(context.Background(), req)
	if err != nil {
		log.Panicln(err)
	}

	log.Println(res.Message)
}
