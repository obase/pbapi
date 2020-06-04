package cache

import (
	"bytes"
	"crypto/md5"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func New(config *Config) Cache {

	config = mergeConfig(config)

	switch strings.ToLower(config.Type) {
	case NONE:
		return newNoneCache(config)
	case REDIS:
		return newRedisCache(config)
	case MEMORY:
		return newMemoryCache(config)
	}
	return nil
}

func ckey(r *http.Request, rbuf *bytes.Buffer) string {

	buffer := buffpool.Get().(*bytes.Buffer)
	defer buffpool.Put(buffer)

	buffer.Reset()
	buffer.WriteString(r.Method)
	buffer.WriteString(":")
	buffer.WriteString(r.URL.Path)
	buffer.WriteString(":")
	buffer.WriteString(r.URL.RawQuery)
	if r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodPut {
		m5 := md5.Sum(rbuf.Bytes())
		buffer.Write(m5[:])
	}
	return buffer.String()
}

func chead(header http.Header) [][]string {
	ret := make([][]string, 0, len(header))
	for k, v := range header {
		ret = append(ret, append(v, k))
	}
	return ret
}

func write(writer gin.ResponseWriter, response *Response) {
	writer.WriteHeader(int(response.Status))
	header := writer.Header()
	for i, n := 0, len(response.Hname); i < n; i++ {
		header[response.Hname[i]] = response.Hvals[i]
	}
	writer.Write(response.Rdata)
}

func read(writer *CacheResponseWriter) *Response {
	resp := new(Response)
	resp.Status = int16(writer.ResponseWriter.Status())
	header := writer.ResponseWriter.Header()
	size := len(header)
	resp.Hname = make([]string, size)
	resp.Hvals = make([][]string, size)
	idx := 0
	for k, v := range header {
		resp.Hname[idx] = k
		resp.Hvals[idx] = make([]string, len(v))
		copy(resp.Hvals[idx], v) //复制形式避免强引用
		idx++
	}
	resp.Rdata = make([]byte, writer.Buffer.Len())
	copy(resp.Rdata, writer.Buffer.Bytes()) // 因为buffer需要重用,此处必须复制
	return resp
}
