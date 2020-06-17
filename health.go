package pbapi

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/obase/center"
	"github.com/obase/log"
	"github.com/obase/pbapi/grpc_health_v1"
	"google.golang.org/grpc"
	"net/http"
	"strconv"
)

const HTTP_HEALTH_PATH = "/health"

func registerServiceHttp(httpServer gin.IRouter, conf *Config) {
	defer log.Flush()
	httpServer.GET(HTTP_HEALTH_PATH, CheckHttpHealth)

	realHttpHost := conf.HttpHost
	if realHttpHost == "" {
		realHttpHost = FirstPrivateAddress
	}

	suffix := "@" + realHttpHost + ":" + strconv.Itoa(conf.HttpPort)
	myname := center.HttpName(conf.Name)
	regs := &center.Service{
		Id:   myname + suffix,
		Kind: "http",
		Name: myname,
		Host: realHttpHost,
		Port: conf.HttpPort,
	}

	chks := &center.Check{
		Type:     "http",
		Target:   fmt.Sprintf("http://%s:%v/health", realHttpHost, conf.HttpPort),
		Timeout:  conf.HttpCheckTimeout,
		Interval: conf.HttpCheckInterval,
	}

	if err := center.Register(regs, chks); err == nil {
		log.Infof("register service success, %v", *regs)
	} else if err != center.ErrInvalidClient {
		log.Errorf("register service error, %v, %v", *regs, err)
	}
}

func registerServiceGrpc(grpcServer *grpc.Server, conf *Config) {

	defer log.Flush()
	service := &HealthService{}
	grpc_health_v1.RegisterHealthServer(grpcServer, service)

	realGrpcHost := conf.GrpcHost
	if realGrpcHost == "" {
		realGrpcHost = FirstPrivateAddress
	}
	suffix := "@" + realGrpcHost + ":" + strconv.Itoa(conf.GrpcPort)
	myname := center.GrpcName(conf.Name)
	regs := &center.Service{
		Id:   myname + suffix,
		Kind: "grpc",
		Name: myname,
		Host: realGrpcHost,
		Port: conf.GrpcPort,
	}
	chks := &center.Check{
		Type:     "grpc",
		Target:   fmt.Sprintf("%s:%v/%v", realGrpcHost, conf.GrpcPort, service),
		Timeout:  conf.GrpcCheckTimeout,
		Interval: conf.GrpcCheckInterval,
	}

	if err := center.Register(regs, chks); err == nil {
		log.Infof("register service success, %v", *regs)
	} else if err != center.ErrInvalidClient {
		log.Errorf("register service error, %v, %v", *regs, err)
	}
}

func deregisterService(conf *Config) {
	// 统一删除
	suffix := "@" + conf.HttpHost + ":" + strconv.Itoa(conf.HttpPort)
	center.Deregister(conf.Name + suffix)
	center.Deregister(center.HttpName(conf.Name) + suffix)
	center.Deregister(center.GrpcName(conf.Name) + suffix)
}

func CheckHttpHealth(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}

type HealthService struct {
}

func (hs *HealthService) Check(context.Context, *grpc_health_v1.HealthCheckRequest) (rsp *grpc_health_v1.HealthCheckResponse, err error) {
	rsp = &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}
	return
}
