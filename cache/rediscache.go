package cache

import (
	"github.com/gin-gonic/gin"
	"github.com/obase/kit"
	"github.com/obase/redis.v2"
	"sync"
)

type redisCache struct {
	*Config
	redis.Redis
	sync.Once
}

func newRedisCache(config *Config) *redisCache {
	return &redisCache{
		Config: config,
	}
}

func (c *redisCache) lazyinit() {
	redis.Setup(&c.Config.Config) // 如果重复则会报duplicate错误,忽略即可!
	c.Redis = redis.Get(c.Config.Config.Key)
}

func (c *redisCache) Cache(seconds int64, f gin.HandlerFunc) gin.HandlerFunc {

	if seconds <= 0 {
		return f
	}

	// 延迟初始化Redis
	c.Once.Do(c.lazyinit)

	// 返回包装句柄
	rdb := c.Redis
	return func(ctx *gin.Context) {
		buf := kit.GetBytesBuffer()
		defer kit.PutBytesBuffer(buf)

		buf.Reset()
		ctx.Request.Body = DupCacheRequestBody(ctx.Request.Body, buf)
		key := ckey(ctx.Request, buf)
		bs, _, _ := redis.Bytes(rdb.Do("GET", key))
		if len(bs) > 0 {
			var rsp Response
			if _, err := rsp.Unmarshal(bs); err == nil {
				write(ctx.Writer, &rsp)
				return
			}
		}

		// 如果没有缓存,则包装writer调用handler
		buf.Reset()
		ctx.Writer = NewCacheResponseWriter(ctx.Writer, buf)
		f(ctx)
		// 只会缓存state位于200~400之间的结果
		if status := ctx.Writer.Status(); status >= c.Config.MinStatusCode && status <= c.Config.MaxStatusCode {
			if bs, err := read(ctx.Writer.(*CacheResponseWriter)).Marshal(nil); err == nil {
				rdb.Do("SETEX", key, seconds, bs)
			}
		}
	}
}

func (c *redisCache) Close() {
	if c.Redis != nil {
		c.Redis.Close()
	}
}
