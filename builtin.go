package pbapi

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func HostsallowServerPlugin(args []string) grpc.ServerOption {
	return nil
}

func HostsallowRouterPlugin(args []string) gin.HandlersChain {
	return nil
}
