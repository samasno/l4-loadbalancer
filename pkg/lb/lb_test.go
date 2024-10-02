package lb

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

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
	res, err := http.Get(addr)
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
