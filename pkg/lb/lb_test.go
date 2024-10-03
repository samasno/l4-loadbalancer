package lb

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestBalancing(t *testing.T) {
	testAddrs := []Host{
		"http://127.0.0.1:5000",
		"http://127.0.0.1:5001",
		"http://127.0.0.1:5002",
		"http://127.0.0.1:5003",
	}

	closer := startTestServers(testAddrs)
	defer closer()

	hm := NewHealthchecker()

	for _, h := range testAddrs {
		err := hm.AddHost(h)
		if err != nil {
			log.Fatalf("health monitor: %s", err.Error())
		}
	}

	um := NewRoundrobin()
	for _, h := range testAddrs {
		err := um.Add(h)
		if err != nil {
			log.Fatalf("round robin: %s", err.Error())
		}
	}

	lb := NewLoadBalancer(hm, um)
	defer lb.Close()

	go func(lb *Layer4LoadBalancer) {
		err := lb.ListenAndServe("127.0.0.1:8080")
		if err != nil && err != net.ErrClosed {
			t.Log(err.Error())
			t.Fail()
		}
	}(lb)

	time.Sleep(time.Duration(20) * time.Second)

	hits := map[Host]int{}

	for _, h := range testAddrs {
		_, ep, _ := normalizeURL(h)
		hits[ep] = 0
	}

	for i := 0; i < 20; i++ {
		ok, msg, err := pingOk("127.0.0.1:8080")
		if err != nil {
			t.Log(err.Error())
			t.Fail()
			break
		}

		if !ok {
			t.Logf("invalid response ping #%d", i)
			t.Fail()
			break
		}

		for k, _ := range hits {
			h, _, _ := normalizeURL(k)
			if strings.Contains(msg, h) {
				hits[k]++
			}
		}
	}

	for k, v := range hits {
		t.Logf("got %d hits on %s", v, k)
		if v == 0 {
			t.Fail()
		}
	}

}

func TestTestServer(t *testing.T) {
	var testAddrs = []string{
		"127.0.0.1:3333",
		"127.0.0.1:3332",
		"127.0.0.1:3331",
	}

	close := startTestServers(testAddrs)

	time.Sleep(time.Duration(1) * time.Second)

	for _, addr := range testAddrs {
		ok, msg, err := pingOk(addr)
		if err != nil {
			t.Log(err.Error())
			t.Fail()
			break
		}

		if !ok {
			t.Logf("")
			t.Fail()
			break
		}

		ok = strings.Contains(msg, addr)
		if !ok {
			t.Logf("test server string \"%s\" not in \"%s\"", addr, msg)
			break
		}
	}

	close()
}

func startTestServers(addrs []string) func() {
	testServers := []*http.Server{}

	for _, addr := range addrs {
		addr, _, _ = normalizeURL(addr)
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

func pingOk(addr string) (bool, string, error) {
	addr = "http://" + addr
	req, _ := http.NewRequest("GET", addr, nil)

	client := http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return false, "", err
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return false, "", errors.New("got a bad status code")
	}

	msg, err := io.ReadAll(res.Body)
	if err != nil {
		return false, "", err
	}

	return true, string(msg), nil
}
