package lb

import (
	"errors"
	"net"

	"github.com/samasno/l4-loadbalancer/pkg/srv"
)

type HealthMonitor interface {
	AddHost(Host, ...string) error
	IsAlive(Host) (bool, error)
}

type UpstreamManager interface {
	Add(Host) error
	Next() (Host, error)
	Peek() Host
	String() string
	Count() int
}

type Host = string

type Layer4LoadBalancer struct {
	connections     map[string]net.Conn
	HealthMonitor   HealthMonitor
	UpstreamManager UpstreamManager
}

func NewLoadBalancer(hm HealthMonitor, u UpstreamManager) *Layer4LoadBalancer {
	l := &Layer4LoadBalancer{
		connections:     map[string]net.Conn{},
		HealthMonitor:   hm,
		UpstreamManager: u,
	}

	return l
}

func (l *Layer4LoadBalancer) Handle(conn net.Conn) {
	l.Balance(conn)
}

func (l *Layer4LoadBalancer) ListenAndServe(addr string) error {
	tcp := srv.NewTcpServer(addr)

	tcp.Handle(l)

	return tcp.ListenAndServe()
}

func (l *Layer4LoadBalancer) Balance(client net.Conn) {
	remote := client.LocalAddr().String()
	l.connections[remote] = client

	var host Host
	c := l.UpstreamManager.Count()
	var err error
	var ok bool

	for i := 0; i < c*2; i++ {
		host, err = l.UpstreamManager.Next()
		if err != nil {
			continue
		}

		ok, err = l.HealthMonitor.IsAlive(host)
		if err != nil {
			continue
		}

		if ok {
			var upstream net.Conn
			host, _, _ := normalizeURL(host)
			upstream, err = connect(host)
			if err != nil {
				continue
			}
			local := upstream.LocalAddr().String()
			l.connections[local] = upstream
			l.pipe(upstream, client)
			return
		} else {
			continue
		}
	}

	if err == nil {
		err = errors.New("HTTP/1.2 500 error\r\nConnection: close\r\nContent-Length: 0")
	}

	client.Write([]byte(err.Error()))
	client.Close()
}

func (l *Layer4LoadBalancer) Close() {
	for _, conn := range l.connections {
		if conn != nil {
			conn.Close()
		}
	}
}

func connect(host string) (net.Conn, error) {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (l *Layer4LoadBalancer) pipe(up, down net.Conn) error {
	if up == nil && down == nil {
		return errors.New("got nil client")
	}

	forwardStream := func(in, out net.Conn) {
		defer func() {
			o := out.LocalAddr().String()
			delete(l.connections, o)

			i := in.LocalAddr().String()
			delete(l.connections, i)

			in.Close()
			out.Close()
		}()

		b := make([]byte, 1024)
		for {
			n, err := in.Read(b)
			if err != nil {
				break
			}

			_, err = out.Write(b[:n])
			if err != nil {
				break
			}
		}
	}

	go forwardStream(up, down)
	go forwardStream(down, up)
	return nil
}
