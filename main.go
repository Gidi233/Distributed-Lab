package main

import (
	"log"
	"time"

	servers "main/server"
)

func main() {
	server := servers.NewRpcServer()

	err := server.Start()
	if err != nil {
		return
	}

	index := 0
	ports := []string{"0.0.0.0:8001", "0.0.0.0:8002", "0.0.0.0:8003", "0.0.0.0:8004", "0.0.0.0:8005"}
	time.Sleep(time.Second)
	for index < 5 {
		go func(i int) {
			consumer := servers.NewConsumer()
			err := consumer.Start(ports[i], int64(i))
			if err != nil {
				log.Println("consumer.Start", i, " err")
				return
			}
		}(index)
		index++
	}

	time.Sleep(time.Second * 100)
}
