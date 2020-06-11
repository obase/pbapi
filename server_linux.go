package pbapi

import (
	"context"
	"github.com/obase/log"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

var flag = os.Getenv(GRACE_ENV)

func graceListenGrpc(host string, port int) (net.Listener, error) {

	if flag != "" {
		var (
			grpcListner net.Listener
			err         error
			fd          uintptr
		)
		switch flag {
		case GRACE_GRPC:
			fd = 3
		case GRACE_ALL:
			fd = 3
		default:
			return nil, nil
		}
		file := os.NewFile(fd, "")
		defer file.Close()
		if grpcListner, err = net.FileListener(file); err != nil {
			log.Error(nil, "FileListener error: %v", err)
		}
		return grpcListner, err
	}
	return net.Listen("tcp", host+":"+strconv.Itoa(port))
}

func graceListenHttp(host string, port int, keepalive time.Duration) (net.Listener, error) {
	if flag != "" {
		var (
			httpListner net.Listener
			err         error
			fd          uintptr
		)
		switch flag {
		case GRACE_HTTP:
			fd = 3
		case GRACE_ALL:
			fd = 4
		default:
			return nil, nil
		}
		file := os.NewFile(fd, "")
		defer file.Close()
		if httpListner, err = net.FileListener(file); err != nil {
			log.Error(nil, "FileListener error: %v", err)
		}
		return httpListner, err
	}

	tln, err := net.Listen("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	return &KeepAliveTCPListener{TCPListener: tln.(*net.TCPListener), KeepAlivePeriod: keepalive}, nil
}

func graceShutdownOrRestart(grpcServer *grpc.Server, grpcListener net.Listener, httpServer *http.Server, httpListener net.Listener) {
	sch := make(chan os.Signal, 1)
	defer signal.Stop(sch)

	signal.Notify(sch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		sig := <-sch

		switch sig {
		case syscall.SIGUSR2:
			var (
				args  []string
				flag  string
				files []*os.File
			)
			// 设置重启标志及参数
			if len(os.Args) > 1 {
				args = os.Args[1:]
			}
			if grpcListener != nil && httpListener != nil {
				flag = GRACE_ALL
				files = []*os.File{GetListenerFile(grpcListener), GetListenerFile(httpListener)}
			} else if grpcListener != nil {
				flag = GRACE_GRPC
				files = []*os.File{GetListenerFile(grpcListener)}
			} else if httpListener != nil {
				flag = GRACE_HTTP
				files = []*os.File{GetListenerFile(httpListener)}
			} else {
				flag = GRACE_NONE
			}

			// 执行重启命令
			cmd := exec.Command(os.Args[0], args...)
			cmd.Env = append(os.Environ(), GRACE_ENV+"="+flag) // 拼加标志
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.ExtraFiles = files
			if err := cmd.Start(); err != nil {
				log.Error(nil, "restart error: %v", err)
			}
			fallthrough

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
