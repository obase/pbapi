package pbapi

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer()
	s.GET("/test", func(context *gin.Context) {
		fmt.Println("this is a test")
	})
	if err := s.Serve(); err != nil {
		t.Fatal(err)
	}
}
