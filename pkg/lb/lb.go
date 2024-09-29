package lb

import "net"

type Layer4LoadBalancer struct {
	conns         []net.Conn
	HealthMonitor string
	Upstreams     string
}

func (l *Layer4LoadBalancer) Balance(client *net.Conn) {
	// get upstream
	// open up connection to upstream
	// pipe between client and upstream

}

func (l *Layer4LoadBalancer) Pipe(up, down net.Conn) {
	go pipe(up, down)
	go pipe(down, up)
}

func pipe(in, out net.Conn) {
	b := make([]byte, 1024)
	for {
		n, err := in.Read(b)
		if err != nil {
			break
		}

		_, err = out.Write(b[:n])
	}
}
