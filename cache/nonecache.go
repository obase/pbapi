package cache

import (
	"github.com/gin-gonic/gin"
)

/*
占位测试的特殊类型
*/
type noneCache struct {
}

func newNoneCache(c *Config) *noneCache {
	return &noneCache{}
}

func (c *noneCache) Cache(seconds int64, f gin.HandlerFunc) gin.HandlerFunc {
	return f
}

func (c *noneCache) Close() {

}
