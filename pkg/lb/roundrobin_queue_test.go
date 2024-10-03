package lb

import (
	"log"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	testHosts := []Host{
		"http://test.com",
		"http://localhost:8080",
		"https://google.com/test",
	}

	r := NewRoundrobin()

	for _, h := range testHosts {
		r.Add(h)
	}

	if len(testHosts) != r.Count() {
		log.Fatalf("expected %d got %d", len(testHosts), r.Count())
	}
}

func TestRoundRobinNext(t *testing.T) {
	testHosts := []Host{
		"http://test.com",
		"http://localhost:8080",
		"https://google.com/test",
	}

	r := NewRoundrobin()

	for _, h := range testHosts {
		r.Add(h)
	}

	for i := 0; i < len(testHosts); i++ {
		next, err := r.Next()
		if err != nil {
			t.Fatalf(err.Error())
		}

		_, h, _ := normalizeURL(testHosts[i])
		if next != h {
			t.Fatalf("expected %s got %s\n", h, next)
		}
	}

}
