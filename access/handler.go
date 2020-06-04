package access

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/obase/log"
	"strconv"
	"sync"
	"time"
)

const SPACE = ' '

func NewHandlerFunc(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		return nil
	}
	var pool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	/*
		- 请求来源
		- 请求方法
		- 请求路径
		- 请求查询串
		- 响应状态
		- 响应时间
	*/
	return func(ctx *gin.Context) {
		start := time.Now().UnixNano()
		source := ctx.ClientIP()
		method := ctx.Request.Method
		path := ctx.Request.URL.Path
		query := ctx.Request.URL.RawQuery
		ctx.Next()
		status := ctx.Writer.Status()
		end := time.Now().UnixNano()
		used := (end - start) / 1000000

		buf := pool.Get().(*bytes.Buffer)
		buf.WriteString(source)
		buf.WriteRune(SPACE)
		buf.WriteString(method)
		buf.WriteRune(SPACE)
		buf.WriteString(path)
		buf.WriteRune(SPACE)
		buf.WriteString(query)
		buf.WriteRune(SPACE)
		buf.WriteString(strconv.Itoa(status))
		buf.WriteRune(SPACE)
		buf.WriteString(strconv.FormatInt(used, 10))
		logger.Info(ctx, buf.String())
		pool.Put(buf)
	}
}
