package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/samasno/l4-loadbalancer/pkg/lb"
)

func main() {
	exampleServers := []lb.Host{
		"http://127.0.0.1:5000",
		"http://127.0.0.1:5001",
	}

	closer := startTestServers(exampleServers)
	defer closer()

	hm := lb.NewHealthchecker()

	for _, h := range exampleServers {
		err := hm.AddHost(h)
		if err != nil {
			log.Fatalf("health monitor: %s", err.Error())
		}
	}

	um := lb.NewRoundrobin()
	for _, h := range exampleServers {
		err := um.Add(h)
		if err != nil {
			log.Fatalf("round robin: %s", err.Error())
		}
	}

	server := lb.NewLoadBalancer(hm, um)
	defer server.Close()

	go func() {
		err := server.ListenAndServe("127.0.0.1:8080")
		if err != nil && err != net.ErrClosed {
			log.Fatal(err.Error())
		}
	}()

	k := make(chan os.Signal, 5)

	signal.Notify(k, syscall.SIGTERM)

	<-k

	server.Close()

	log.Println("closed server")
}

func startTestServers(addrs []string) func() {
	testServers := []*http.Server{}

	for _, addr := range addrs {
		addr := strings.Replace(addr, "http://", "", 1)
		srv := newTestServer(addr)
		testServers = append(testServers, srv)
		go func(srv *http.Server) {
			err := srv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Printf("test server %s: %s", addr, err.Error())
			}
		}(srv)
	}

	closeServers := func() {
		for _, srv := range testServers {
			srv.Close()
		}
	}

	return closeServers
}

func newTestServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("home page %s", addr)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(msg))
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return srv
}
