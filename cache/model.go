package cache

import (
	"bufio"
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	"sync"
)

const (
	NONE   string = ""
	REDIS  string = "redis"
	MEMORY string = "memory"

	BufferBlockSize = 10240 // 10k
)

type Cache interface {
	Cache(seconds int64, h gin.HandlerFunc) gin.HandlerFunc
	Close()
}

type CacheRequestBody bytes.Buffer

func DupCacheRequestBody(body io.ReadCloser, buffer *bytes.Buffer) *CacheRequestBody {
	defer body.Close()

	buffer.Reset()
	io.Copy(buffer, bufio.NewReader(body))
	return (*CacheRequestBody)(buffer)
}

func (b *CacheRequestBody) Read(p []byte) (n int, err error) {
	return (*bytes.Buffer)(b).Read(p) // 必须转换成bytes.Buffer, 否则会变成自递归, 抛overflow
}

func (b *CacheRequestBody) Close() error {
	return nil
}

type CacheResponseWriter struct {
	gin.ResponseWriter
	Buffer *bytes.Buffer
}

func NewCacheResponseWriter(writer gin.ResponseWriter, buffer *bytes.Buffer) *CacheResponseWriter {
	return &CacheResponseWriter{
		ResponseWriter: writer,
		Buffer:         buffer,
	}
}

func (w *CacheResponseWriter) Write(data []byte) (int, error) {
	w.Buffer.Write(data)
	return w.ResponseWriter.Write(data)
}

var buffpool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, BufferBlockSize)) // 默认10K
	},
}
