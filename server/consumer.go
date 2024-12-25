package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"main/kitex_gen/api"
	"main/kitex_gen/api/client_operations"
	"main/kitex_gen/api/server_operations"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/server"
)

type Consumer struct {
	me   int64
	port string
	cli  *client_operations.Client
}

func NewConsumer() Consumer {
	return Consumer{}
}

func (c *Consumer) Start(ports string, index int64) error {

	addr_cli, _ := net.ResolveTCPAddr("tcp", ports)

	var opts []server.Option

	opts = append(opts, server.WithServiceAddr(addr_cli), server.WithReusePort(true))

	srv_bro_cli := server_operations.NewServer(c, opts...)

	go func() {
		err := srv_bro_cli.Run()

		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	c.me = index
	c.port = ports

	cli, err := client_operations.NewClient("consumer", client.WithHostPorts("0.0.0.0:7778"))
	if err != nil {
		return err
	}
	c.cli = &cli

	for in := 0; in < 10; in++ {
		resp, err := cli.Request(context.Background(), &api.RequestAsgs{
			Consumer: c.me,
			Port:     c.port,
		})

		if err != nil {
			fmt.Println(err.Error())
			return err
		} else if resp.Ret { //成功请求到资源，使用完进行释放
			fmt.Println("request resource success")

			time.Sleep(time.Second * 3)

			ret, _ := cli.Release(context.Background(), &api.ReleaseAsgs{
				Consumer: c.me,
			})

			if ret.Ret {
				fmt.Println("Release resource success")
			} else {
				fmt.Println("Release resource fail")
			}
		} else {
			fmt.Println("request resource fail to wait")
		}
	}

	return nil
}

func (c *Consumer) Reply(ctx context.Context, req *api.ReplyArgs_) (resp *api.ReplyReply, err error) {

	fmt.Println("the resource requested")

	var res api.ReplyReply

	res.Ret = true

	go func() {
		time.Sleep(time.Second * 3)

		resp, _ := (*c.cli).Release(context.Background(), &api.ReleaseAsgs{
			Consumer: c.me,
		})

		if resp.Ret {
			fmt.Println("Release resource success")
		} else {
			fmt.Println("Release resource fail")
		}

	}()

	return &res, nil
}
