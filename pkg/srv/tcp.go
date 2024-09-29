package srv

import "net"

type TCP struct {
	Addr    string
	ln      net.Listener
	conns   []net.Conn
	handler func(*net.Conn)
}

// have single handler for server
type TcpHandler func(*net.Conn)

func (t *TCP) HandleFunc(fn TcpHandler) {
	t.handler = fn
}

func NewTcpServer(addr string) *TCP {
	s := &TCP{
		Addr: addr,
	}

	return s
}

// blocking
func (t *TCP) ListenAndServe() error {
	ln, err := net.Listen("tcp", t.Addr)
	if err != nil {
		return err
	}

	t.ln = ln

	for {
		conn, err := t.ln.Accept()
		if err != nil {
			if err == net.ErrClosed {
				break
			}
		}

		go t.handler(&conn)
	}

	return nil
}

func (t *TCP) Shutdown() error {
	return t.ln.Close()
}
