package pbapi

import (
	"net"
	"os"
	"time"
)

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type KeepAliveTCPListener struct {
	*net.TCPListener
	KeepAlivePeriod time.Duration
}

func (ln KeepAliveTCPListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	if ln.KeepAlivePeriod == 0 {
		tc.SetKeepAlivePeriod(3 * time.Minute)
	} else {
		tc.SetKeepAlivePeriod(ln.KeepAlivePeriod)
	}
	return tc, nil
}

func GetListenerFile(l net.Listener) *os.File {
	var file *os.File
	switch l := l.(type) {
	case *KeepAliveTCPListener:
		file, _ = l.TCPListener.File()
	case *net.TCPListener:
		file, _ = l.File()
	}
	return file
}

var FirstPrivateAddress = func(def string) (ret string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return def
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				// 必须私有网段
				if (ip4[0] == 10) || (ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || (ip4[0] == 192 && ip4[1] == 168) {
					return ip4.String()
				}
			}
		}
	}
	return def
}("127.0.0.1")
