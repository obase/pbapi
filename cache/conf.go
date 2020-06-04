package cache

import (
	"github.com/obase/redis"
	"math"
)

type Config struct {
	redis.Config  `bson:",inline" yaml:",inline"`
	Type          string `json:"type" bson:"type" yaml:"type"`
	MaxMemorySize int    `json:"maxMemorySize" bson:"maxMemorySize" yaml:"maxMemorySize"`
	MinStatusCode int    `json:"minStatusCode" bson:"minStatusCode" yaml:"minStatusCode"`
	MaxStatusCode int    `json:"maxStatusCode" bson:"maxStatusCode" yaml:"maxStatusCode"`
}

func mergeConfig(config *Config) *Config {
	if config == nil {
		config = new(Config)
	}

	if config.MaxMemorySize == 0 {
		config.MaxMemorySize = math.MaxInt16
	}

	if config.MinStatusCode == 0 {
		config.MinStatusCode = 200
	}

	if config.MaxStatusCode == 0 {
		config.MaxStatusCode = 399
	}

	return config
}
