package server

import (
	"context"
	"fmt"
	"main/kitex_gen/api"
	"main/kitex_gen/api/node_operations"
	"net"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/server"
)

type Node struct {
	Id int64 //作为优先级
	Ip string
	node_operations.Client
}
type empty struct{}

type RPCServer struct {
	mu       sync.Mutex
	id       int64 //作为优先级
	ip       string
	isGet    bool
	resource atomic.Bool //协调者临界资源，true标识资源可用，false表示不可用。
	// otherElection atomic.Bool //true标识选举中，false表示选举结束。
	otherElection chan empty //true标识选举中，false表示选举结束。

	requests []int64 //id等待队列
	leader   *Node

	cliMu   sync.RWMutex
	clients []*Node  //id优先级排序
	ip2Node sync.Map //IP->node
}

func NewRpcServer() RPCServer {
	return RPCServer{
		mu: sync.Mutex{},
	}
}

func (s *RPCServer) PutLeader() {
	fmt.Println("leader:", s.leader)

	fmt.Println("list:")
	for _, n := range s.clients {
		fmt.Println(n)
	}
}

func (s *RPCServer) Start(id int64, curNode Node, nodes []Node) error {
	s.ip = curNode.Ip
	s.id = id
	addr_cli, _ := net.ResolveTCPAddr("tcp", curNode.Ip)
	var opts []server.Option

	opts = append(opts, server.WithServiceAddr(addr_cli), server.WithReusePort(true))

	srv := node_operations.NewServer(s, opts...)
	s.resource.Store(true)
	fmt.Printf("start coordinator server\n")
	go func() {
		err := srv.Run()

		if err != nil {
			fmt.Println(err.Error())
		}
	}()
	for i, n := range nodes {
		fmt.Println(n)
		if n.Ip == s.ip {
			continue
		}
		go func(node Node, i int) {
			for {
				con_cli, err := node_operations.NewClient("node"+strconv.Itoa(i), client.WithHostPorts(node.Ip)) //Node_Operations
				if err != nil {
					time.Sleep(time.Second)
					fmt.Printf("connection error: %v\n", err)
					continue
				}
				node.Client = con_cli
				fmt.Println("map store :", node.Ip, node)
				s.ip2Node.Store(node.Id, &node)
				s.clients = append(s.clients, &node)
				sort.Slice(s.clients, func(i, j int) bool {
					return s.clients[i].Id < s.clients[j].Id
				})
				return
			}
		}(n, i)
	}

	if ok := s.startElection(); !ok {
		fmt.Println("Election failed: ")
	}

	return nil
}

func (s *RPCServer) Request(ctx context.Context, req *api.RequestAsgs) (resp *api.RequestReply, err error) {

	id := req.Id
	fmt.Println("Request from: ", id)
	s.mu.Lock()
	defer s.mu.Unlock()

	var res api.RequestReply
	// fmt.Println("old resource:", s.resource.Load())
	if s.resource.CompareAndSwap(true, false) { //如果有资源，分配
		res.Ret = true // 	s.ip2Node[req.Consumer] = &con_cli

	} else { //没有资源，将请求对象加入requests队列中，等待有资源
		res.Ret = false

		s.requests = append(s.requests, id)
	}
	// fmt.Println("new resource:", s.resource.Load())
	// fmt.Println("res:", res)

	return &res, nil
}

func (s *RPCServer) Release(ctx context.Context, req *api.ReleaseAsgs) (resp *api.ReleaseReply, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Println("Release from: ", req.Id)
	var res api.ReleaseReply

	s.helpReply()
	res.Ret = true

	return &res, nil
}

func (s *RPCServer) helpReply() {
	if len(s.requests) > 0 { //还有正在等待的request，分配资源
		// i:=s.requests[0]
		if s.requests[0] != s.id {
			n, ok := s.ip2Node.Load(s.requests[0])
			if !ok {
				fmt.Println(" requests slice:", s.requests)
				fmt.Printf("coordinator server not connection consumer %v\n", s.requests[0])
			} else {
				node := n.(*Node)
				reply, err := node.Reply(context.Background(), &api.ReplyArgs_{
					Id: s.id,
				})
				if err != nil || !reply.Ret {
					fmt.Printf("coordinator server reply fail %v\n", err)
					return
				}
			}
		} else {
			fmt.Println("the resource requested")
		}
		s.requests = s.requests[1:] //处理过的放出来
	} else {
		s.resource.Store(true) //释放资源
	}
}

func (s *RPCServer) Reply(ctx context.Context, req *api.ReplyArgs_) (resp *api.ReplyReply, err error) {
	s.isGet = true
	fmt.Println("the resource requested")

	var res api.ReplyReply

	res.Ret = true

	// go func() {
	// 	time.Sleep(time.Second * 3)

	// 	resp, _ := s.leader.Release(context.Background(), &api.ReleaseAsgs{
	// 		Consumer: s.id,
	// 	})

	// 	if resp.Ret {
	// 		fmt.Println("Release resource success")
	// 	} else {
	// 		fmt.Println("Release resource fail")
	// 	}

	// }()

	return &res, nil
}

// 操作多了会收不到其他节点的调用(好像调用一次就调不了了)
func (s *RPCServer) Election(ctx context.Context, req *api.ElectionAsgs) (resp *api.ElectionReply, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := int64(req.Id)
	// 获取发起选举请求的节点信息
	fmt.Println("recieve from:", id)
	n, ok := s.ip2Node.Load(id)
	sourceNode := n.(*Node)
	if !ok {
		return &api.ElectionReply{
			Ret: false,
		}, nil
	}

	// 如果请求节点优先级低于当前节点，拒绝其成为协调者
	if sourceNode.Id < s.id {

		s.otherElection = make(chan empty,5)
		go s.startElection()

		return &api.ElectionReply{
			Ret: false,
		}, nil
	}

	// 如果请求节点优先级更高，承认其为协调者
	s.leader = sourceNode
	fmt.Println("change leader to", id, "new", s.leader)
	// 加鎖{
	if s.otherElection != nil {
		s.otherElection <- empty{} //会阻塞？ true(这种并发响应不应该直接起协程吗，为什么一种方法会阻塞所有的调用)
	}
// }
	fmt.Println("Election return true")

	return &api.ElectionReply{
		Ret: true,
	}, nil
}

// 启动选举流程
func (s *RPCServer) startElection() bool {
	fmt.Println("start election one term")
	// s.otherElection.Store(true)
	responses := make(chan *api.ElectionReply, len(s.clients))
	timeout := time.After(time.Second * 5) // T时间

	higherNum := 0
	// 向更高优先级的节点发送选举请求
	for i := len(s.clients) - 1; i > int(s.id); i-- {
		higherNum++
		node := s.clients[i]
		fmt.Println("send election to", node)
		go func(n *Node) {
			resp, err := n.Election(context.Background(), &api.ElectionAsgs{
				Id: s.id,
			})
			if err == nil {
				responses <- resp
			} else {
				fmt.Printf("invoke Election failed: %v", err)
			}
		}(node)
	}

	// 等待响应
	responseCount := 0
	for {
		select {
		case resp := <-responses:
			responseCount++
			if !resp.Ret {
				select {
				case <-time.After(time.Second * 5): // T'时间
					s.otherElection = nil
					// 重新开始选举
					return s.startElection()
				case <-s.otherElection:
					s.otherElection = nil
					return false
				}
			}
			// 如果收到了所有响应且没有人反对
			if responseCount == higherNum {
				s.becomeLeader()
				return true
			}

		case <-timeout:
			s.becomeLeader()
			return true
		}
	}
}

// 成为协调者
func (s *RPCServer) becomeLeader() {
	fmt.Println("become leader")
	s.leader = &Node{
		Id: s.id,
		Ip: s.ip,
	}

	// 通知所有其他节点
	fmt.Println("before leader", s.clients)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var wg sync.WaitGroup
	for _, node := range s.clients {
		wg.Add(1) // 为每个 goroutine 添加一个计数
		go func(n *Node) {
			defer wg.Done() // 每个 goroutine 完成后调用 Done()
			fmt.Println("node:", *n)
			ret, err := n.Election(ctx, &api.ElectionAsgs{
				Id: s.id,
			})
			fmt.Println("election resp:", ret, "  .err:", err)
		}(node)
	}

	wg.Wait()

	fmt.Println("all election responses received")
}

func (s *RPCServer) Get() error {
	for {
		if s.isGet {
			return fmt.Errorf("resource has already request")
		}

		if s.leader.Ip == s.ip {
			if !s.resource.CompareAndSwap(true, false) { //没有资源，将请求对象加入requests队列中，等待有资源
				fmt.Println("cur add", s.id)
				s.requests = append(s.requests, s.id)
				return fmt.Errorf("resource not available, request queued")
			} else {
				s.isGet = true
				fmt.Printf("Successfully acquired resource\n")
				return nil
			}
		}

		// 发送资源请求
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		resp, err := s.leader.Request(ctx, &api.RequestAsgs{
			Id:   s.id,
		})

		if err != nil {
			fmt.Printf("Election failed: %v", err)
			// 如果请求失败，可能是leader出问题，重新选举

			s.leader = nil
			if ok := s.startElection(); !ok {
				fmt.Println("Election failed: ")
				continue
			} else {
				s.resource.Store(true)
				return nil
			}
		}

		if !resp.Ret {
			return fmt.Errorf("resource not available, request queued")
		}
		s.isGet = true
		fmt.Printf("Successfully acquired resource\n")
		return nil
	}
}

func (s *RPCServer) Put() error {
	fmt.Println("isget", s.isGet)
	if !s.isGet {
		return fmt.Errorf("resource not acquired")
	}

	if s.leader == nil {
		return fmt.Errorf("no leader available")
	}

	if s.leader.Ip == s.ip {
		s.helpReply()
		fmt.Printf("Successfully released resource\n")
		return nil
	}
	// 发送释放资源请求
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := s.leader.Release(ctx, &api.ReleaseAsgs{
		Id: s.id,
	})

	if err != nil {
		fmt.Printf("Put failed: %v", err)
		s.leader = nil
		// election
		return fmt.Errorf("failed to release resource: %v", err)
	}

	if !resp.Ret {
		return fmt.Errorf("failed to release resource")
	}
	s.isGet = false
	fmt.Printf("Successfully released resource\n")
	return nil
}
