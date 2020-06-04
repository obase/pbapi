package cache

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"sync"
	"time"
)

type memoryEntry struct {
	Time int64
	*Response
}

type memoryCache struct {
	*Config
	sync.RWMutex
	Data map[string]*memoryEntry
}

func newMemoryCache(config *Config) *memoryCache {
	return &memoryCache{
		Config: config,
		Data:   make(map[string]*memoryEntry),
	}
}

func (c *memoryCache) Cache(seconds int64, f gin.HandlerFunc) gin.HandlerFunc {

	if seconds <= 0 {
		return f
	}

	return func(ctx *gin.Context) {
		buf := buffpool.Get().(*bytes.Buffer)
		defer buffpool.Put(buf)

		buf.Reset()
		ctx.Request.Body = DupCacheRequestBody(ctx.Request.Body, buf)
		key := ckey(ctx.Request, buf)

		c.RWMutex.RLock()
		entry, ok := c.Data[key]
		c.RWMutex.RUnlock()

		now := time.Now().Unix()
		if ok && now-entry.Time < seconds {
			write(ctx.Writer, entry.Response)
			return
		}

		buf.Reset()
		ctx.Writer = NewCacheResponseWriter(ctx.Writer, buf)
		f(ctx)
		// 只会缓存state位于200~400之间的结果
		if status := ctx.Writer.Status(); status >= c.Config.MinStatusCode && status <= c.Config.MaxStatusCode {
			if entry == nil {
				entry = new(memoryEntry)
				// 理论上面entry也是需要同步控制,为了性能此处舍弃!
				entry.Time = now
				entry.Response = read(ctx.Writer.(*CacheResponseWriter))
				c.RWMutex.Lock()
				if len(c.Data) >= c.MaxMemorySize {
					c.Data = make(map[string]*memoryEntry) // 重新释放
				}
				c.Data[key] = entry
				c.RWMutex.Unlock()
			} else {
				// 理论上面entry也是需要同步控制,为了性能此处舍弃!
				entry.Time = now
				entry.Response = read(ctx.Writer.(*CacheResponseWriter))
			}
		}
	}
}

func (c *memoryCache) Close() {
	c.RWMutex.Lock()
	c.Data = nil
	c.RWMutex.Unlock()
}
