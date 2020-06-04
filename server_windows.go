package pbapi

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func graceListenGrpc(host string, port int) (net.Listener, error) {
	return net.Listen("tcp", host+":"+strconv.Itoa(port))
}

func graceListenHttp(host string, port int, keepalive time.Duration) (net.Listener, error) {
	tln, err := net.Listen("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	return &KeepAliveTCPListener{TCPListener: tln.(*net.TCPListener), KeepAlivePeriod: keepalive}, nil
}

func graceShutdownOrRestart(grpcServer *grpc.Server, grpcListener net.Listener, httpServer *http.Server, httpListener net.Listener) {
	sch := make(chan os.Signal, 1)
	defer signal.Stop(sch)

	signal.Notify(sch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		sig := <-sch

		switch sig {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
			ws := new(sync.WaitGroup)
			if httpServer != nil {
				ws.Add(1)
				go func(ws *sync.WaitGroup) {
					defer ws.Done()
					httpServer.Shutdown(context.Background())
				}(ws)
			}
			if grpcServer != nil {
				ws.Add(1)
				go func(ws *sync.WaitGroup) {
					defer ws.Done()
					grpcServer.GracefulStop()
				}(ws)
			}
			ws.Wait()
			return
		}
	}
}
