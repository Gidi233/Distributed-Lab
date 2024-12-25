package server

import (
	"context"
	"fmt"
	"main/kitex_gen/api"
	"main/kitex_gen/api/client_operations"
	"main/kitex_gen/api/server_operations"
	"net"
	"sync"
	"sync/atomic"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/server"
)

type RPCServer struct {
	mu sync.Mutex

	resource atomic.Bool //临界资源，true标识资源可用，false表示不可用。

	requests []int64 //等待队列

	consumers map[int64]*server_operations.Client //consumerIndex->request conn
}

func NewRpcServer() RPCServer {
	return RPCServer{
		mu:        sync.Mutex{},
		consumers: make(map[int64]*server_operations.Client),
	}
}

func (s *RPCServer) Start() error {

	addr_cli, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:7778")
	var opts []server.Option

	opts = append(opts, server.WithServiceAddr(addr_cli), server.WithReusePort(true))

	srv_bro_cli := client_operations.NewServer(s, opts...)
	s.resource.Store(true)
	fmt.Printf("start coordinator server\n")
	go func() {
		err := srv_bro_cli.Run()

		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	return nil
}

func (s *RPCServer) Request(ctx context.Context, req *api.RequestAsgs) (resp *api.RequestReply, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res api.RequestReply
	// fmt.Println("old resource:", s.resource.Load())
	if s.resource.CompareAndSwap(true, false) { //如果有资源，分配
		res.Ret = true

		// s.resource = false
	} else { //没有资源，将请求对象加入requests队列中，等待有资源
		res.Ret = false

		_, ok := s.consumers[req.Consumer]
		if !ok {
			con_cli, err := server_operations.NewClient("consumer", client.WithHostPorts(req.Port))
			if err != nil {
				return nil, err
			}
			s.consumers[req.Consumer] = &con_cli
		}

		s.requests = append(s.requests, req.Consumer)
	}
	// fmt.Println("new resource:", s.resource.Load())
	// fmt.Println("res:", res)

	return &res, nil
}

func (s *RPCServer) Release(ctx context.Context, req *api.ReleaseAsgs) (resp *api.ReleaseReply, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res api.ReleaseReply

	s.resource.Store(true) //释放资源

	if len(s.requests) > 0 { //还有正在等待的request，分配资源
		con_cli, ok := s.consumers[s.requests[0]]

		if !ok {
			fmt.Printf("coordinator server not connection consumer %v\n", s.requests[0])
		} else {
			reply, err := (*con_cli).Reply(context.Background(), &api.ReplyArgs_{
				Resource: true,
			})
			if err != nil || !reply.Ret {
				s.resource.Store(false)
			} else {
				s.requests = s.requests[1:] //处理过的放出来
			}
		}
	}

	res.Ret = true

	return &res, nil
}
