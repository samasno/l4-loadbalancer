package lb

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type RoundrobinQueue struct {
	pointer int
	hosts   []Host
	mtx     *sync.Mutex
}

func NewRoundrobin() *RoundrobinQueue {
	r := &RoundrobinQueue{
		pointer: 0,
		hosts:   []Host{},
		mtx:     &sync.Mutex{},
	}

	return r
}

type Host = string

func (r *RoundrobinQueue) Enqueue(host Host) error {
	_, host, err := normalizeURL(host)
	if err != nil {
		return err
	}

	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.hosts = append(r.hosts, host)

	return nil
}

func (r *RoundrobinQueue) Dequeue() (Host, error) {
	if len(r.hosts) == 0 {
		return "", errors.New("queue is empty")
	}

	return r.next(), nil
}

func (r *RoundrobinQueue) Peek() Host {
	if len(r.hosts) == 0 {
		return ""
	}

	return r.hosts[r.pointer]
}

func (r *RoundrobinQueue) next() Host {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	host := r.hosts[r.pointer]
	r.pointer = (r.pointer + 1) % len(r.hosts)

	return host
}

func (r *RoundrobinQueue) String() string {
	out := strings.Builder{}

	out.WriteString(fmt.Sprintf("Count: %d\n", len(r.hosts)))
	out.WriteString("Hosts:\n")

	p := r.pointer

	for i := 0; i < len(r.hosts); i++ {
		out.WriteString(fmt.Sprintf("%s;\n", r.hosts[p]))
		p++
	}

	return out.String()
}
