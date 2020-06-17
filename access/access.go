package access

import (
	"github.com/gin-gonic/gin"
	"github.com/obase/kit"
	"github.com/obase/log"
	"strconv"
	"time"
)

const SPACE = ' '

func NewHandlerFunc(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		return nil
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

		buf := kit.GetStringBuffer()
		buf.WriteString(source)
		buf.WriteByte(SPACE)
		buf.WriteString(method)
		buf.WriteByte(SPACE)
		buf.WriteString(path)
		buf.WriteByte(SPACE)
		buf.WriteString(query)
		buf.WriteByte(SPACE)
		buf.WriteString(strconv.Itoa(status))
		buf.WriteByte(SPACE)
		buf.WriteString(strconv.FormatInt(used, 10))
		logger.Info(buf.String())
		kit.PutStringBuffer(buf)
	}
}
