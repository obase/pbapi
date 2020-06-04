package access

import (
	"context"
	"fmt"
	"github.com/obase/log"
	"os"
	"time"
)

type Config struct {
	FlushPeriod     time.Duration `json:"flushPeriod" bson:"flushPeriod" yaml:"flushPeriod"`
	Path            string        `json:"path" bson:"path" yaml:"path"`
	RotateBytes     int64         `json:"rotateBytes" bson:"rotateBytes" yaml:"rotateBytes"`
	RotateCycle     string        `json:"rotateCycle" bson:"rotateCycle" yaml:"rotateCycle"`             //轮转周期,目前仅支持
	BufioWriterSize int           `json:"bufioWriterSize" bson:"bufioWriterSize" yaml:"bufioWriterSize"` //Buffer写缓存大小
}

func NewLogger(ctx context.Context, c *Config) (ret *log.Logger, err error) {

	if c == nil || c.Path == "" {
		return
	}

	ret, err = log.NewBuiltinLogger(&log.Config{
		Level:           log.INFO,
		Path:            c.Path,
		RotateBytes:     c.RotateBytes,
		RotateCycle:     log.GetCycle(c.RotateCycle),
		BufioWriterSize: c.BufioWriterSize,
	})

	// 后台启动flush
	go func(flushInterval time.Duration) {
		if flushInterval <= 0 {
			flushInterval = log.DEFAULT_FLUSH_PERIOD
		}
		for {
			tick := time.Tick(flushInterval)
			select {
			case <-ctx.Done():
				return
			case <-tick:
				protect(ret.Flush)
			}
		}

	}(c.FlushPeriod)

	return
}
func protect(fn func()) {
	defer func() {
		if perr := recover(); perr != nil {
			fmt.Fprintf(os.Stderr, "access flush panic: %v", perr)
		}
	}()
	fn()
}
