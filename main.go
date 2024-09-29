package main

import (
	"github.com/samasno/l4-loadbalancer/pkg/srv"
	"log"
	"net"
	"os"
)

func main() {
	server := srv.NewTcpServer(":8080")

	if err := server.ListenAndServe(); err != nil {
		if err == net.ErrClosed {
			log.Println("server closed")
			os.Exit(0)
		}

		log.Fatal(err.Error())
	}
}
