package lb

import (
	"net/http"
	"testing"
)

func testServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("home page :%d", addr)
		w.Write([]byte(msg))
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return srv
}

func TestTestServer(t *testing.T) {

}
