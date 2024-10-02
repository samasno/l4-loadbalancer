package lb

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HealthChecker map[Host]*healthStatus

func NewHealthchecker() *HealthChecker {
	return &HealthChecker{}
}

func (hc HealthChecker) IsAlive(host Host) (bool, error) {
	h, _, err := normalizeURL(host)

	if err != nil {
		return false, err
	}

	hs, ok := hc[h]
	if !ok {
		return false, nil
	}

	return hs.IsAlive(), nil
}

func (hc HealthChecker) AddHost(host Host, path ...string) error {
	h, endpoint, err := normalizeURL(host, path...)
	if err != nil {
		return err
	}

	hs := &healthStatus{
		alive:    false,
		endpoint: endpoint,
		failed:   0,
		passed:   0,
	}

	hc[h] = hs

	go hs.RunCheck(5)

	return nil
}

func (hc HealthChecker) String() string {
	out := strings.Builder{}
	for _, hs := range hc {
		out.WriteString(hs.String() + "\n")
	}

	return out.String()
}

func (hc HealthChecker) Info(host Host) string {
	hs, ok := hc[host]
	if !ok {
		return ""
	}

	return hs.String()
}

type healthStatus struct {
	alive    bool
	endpoint string
	failed   int
	passed   int
}

func (hs *healthStatus) IsAlive() bool {
	return hs.passed >= 2
}

func (hs *healthStatus) RunCheck(interval int) {
	for {
		hs.check()
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func (hs *healthStatus) check() {
	r, err := http.Get(hs.endpoint)
	if err != nil {
		hs.failed++
		hs.passed = 0

		if hs.failed == 3 {
			hs.alive = false
		}

		return
	}

	defer r.Body.Close()

	if r.StatusCode < 400 {
		hs.failed = 0
		hs.passed++
	} else {
		hs.passed = 0
		hs.failed++
	}

	if hs.passed == 2 {
		hs.alive = true
		return
	}

	if hs.failed == 3 {
		hs.alive = false
		return
	}

}

func (hs *healthStatus) String() string {
	return fmt.Sprintf("endpoint: %s; alive: %v; passed: %d; failed: %d;", hs.endpoint, hs.alive, hs.passed, hs.failed)
}

func normalizeURL(host string, path ...string) (Host, string, error) {
	p, err := url.JoinPath(host, path...)
	if err != nil {
		return "", "", err
	}

	u, err := url.Parse(p)
	if err != nil {
		return "", "", err
	}

	return u.Host, u.String(), nil
}
