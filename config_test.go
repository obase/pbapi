package pbapi

import (
	"encoding/json"
	"fmt"
	"github.com/obase/conf"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	c := LoadConfig()
	if c == nil {
		t.Log("conf.yml is empty")
	} else {
		bs, err := json.Marshal(c.ServerConfig)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", bs)
	}
}

func TestExt(t *testing.T)  {
	c, _ := conf.GetMap("ext")
	for k, v := range c {
		fmt.Println(k, "=>", v)
	}
}
