package lb

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestHealthCheckerAddHost(t *testing.T) {
	testHosts := []Host{
		"http://test.com",
		"http://example.com",
		"https://google.com",
	}

	testPath := "/healthz"

	var expectedB strings.Builder

	for _, h := range testHosts {
		p, _ := url.JoinPath(string(h), testPath)
		expectedB.WriteString(fmt.Sprintf("endpoint: %s; alive: false; passed: 0; failed: 0;\n", p))
	}

	hc := NewHealthchecker()

	for _, h := range testHosts {
		hc.AddHost(h, testPath)
	}

	expected := expectedB.String()
	got := hc.String()

	if expected != got {
		t.Fatalf("expected \n %s got\n %s\n", expected, got)
	}

}

func TestCheckingServerHealth(t *testing.T) {
	liveServers := []string{
		"localhost:8332",
		"localhost:8331",
	}

	deadServers := []string{
		"locahost:8334",
		"localhost:8335",
	}

	allservers := append(liveServers, deadServers...)
	hosts := []Host{}
	for _, h := range allservers {
		hosts = append(hosts, Host("http://"+h))
	}

	closer := startTestServers(liveServers)
	defer closer()

	hc := NewHealthchecker()

	for _, h := range hosts {
		hc.AddHost(h)
	}

	time.Sleep(time.Duration(20) * time.Second)

	for _, h := range hosts[:2] {
		ok, _ := hc.IsAlive(h)
		if !ok {
			fmt.Printf("%s should be alive\n", h)
			t.Fail()
		}
	}

	for _, h := range hosts[2:] {
		ok, _ := hc.IsAlive(h)
		if ok {
			fmt.Printf("%s should be dead\n", h)
			t.Fail()
		}
	}

}

func TestNormalizeHost(t *testing.T) {
	h, e, err := normalizeURL("http://localhost:8080", "test/test")
	if err != nil {
		t.Fatal(err.Error())
	}

	println(h, e)
}
