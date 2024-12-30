package main

import (
	"bufio"
	"fmt"
	"main/server"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <port>")
		return
	}

	port := os.Args[1]

	svr := server.NewRpcServer()

	nodes := []server.Node{//todo:元数据从etcd获取
		{Ip: "127.0.0.1:8001", Id: 1},
		{Ip: "127.0.0.1:8002", Id: 2},
		{Ip: "127.0.0.1:8003", Id: 3},
		// {Ip: "127.0.0.1:8004", Id: 4},
		// {Ip: "127.0.0.1:8005", Id: 5},
	}
	currentIp := "127.0.0.1:" + port
	node := server.Node{Ip: currentIp, Id: func() int64 {
		for _, n := range nodes {
			if n.Ip == currentIp {
				return n.Id
			}
		}
		return 0 
	}()}
	err := svr.Start(node.Id, node, nodes)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		return
	}

	// 创建用户输入处理的扫描器
	scanner := bufio.NewScanner(os.Stdin)

	// 用户交互循环
	for {
		fmt.Println("\nEnter command:")
		fmt.Println("1: Request resource (GET)")
		fmt.Println("2: Release resource (PUT)")
		fmt.Println("0: Exit")
		fmt.Print("Enter choice: ")

		// 读取用户输入
		scanner.Scan()
		input := scanner.Text()

		switch input {
		case "1":
			// 调用 Get 方法
			err := svr.Get()
			if err != nil {
				fmt.Printf("Error in GET request: %v\n", err)
			} else {
				fmt.Println("Resource successfully acquired.")
			}
		case "2":
			// 调用 Put 方法
			err := svr.Put()
			if err != nil {
				fmt.Printf("Error in PUT request: %v\n", err)
			} else {
				fmt.Println("Resource successfully released.")
			}
		case "9":
			svr.PutLeader()
		case "0":
			// 退出程序
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}

		// 给操作留一些时间进行处理
		time.Sleep(time.Millisecond * 100)
	}
}
