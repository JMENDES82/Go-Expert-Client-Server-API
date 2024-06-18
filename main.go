package main

import (
	"fmt"
	"time"

	"go-expert/client"
	"go-expert/server"
)

func main() {
	go func() {
		server.Server()
	}()

	time.Sleep(2 * time.Second)

	client.Client()

	fmt.Println("Aplicação client e server executadas com sucesso")
}
