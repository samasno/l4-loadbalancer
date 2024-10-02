package lb

import "net"

type Layer4LoadBalancer struct {
	conns           map[string]net.Conn
	HealthMonitor   string
	UpstreamManager string
}

func (l *Layer4LoadBalancer) Balance(client *net.Conn) {
	// get upstream
	// open up connection to upstream
	// strip protocol
	// pipe between client and upstream
}

func (l *Layer4LoadBalancer) Pipe(up, down net.Conn) {
	go pipe(up, down)
	go pipe(down, up)
}

func pipe(in, out net.Conn) {
	defer in.Close()
	defer out.Close()

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
