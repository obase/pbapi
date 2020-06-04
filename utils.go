package pbapi

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/obase/log"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const (
	GRACE_ENV  = "_GRC_"
	GRACE_NONE = "0"
	GRACE_GRPC = "1"
	GRACE_HTTP = "2"
	GRACE_ALL  = "3" // grpc是3, http是4
)

var JsonContentType = []string{"application/json; charset=utf-8"}

/*
支持*与？的通配规则：
1. *表示任意字符
2. ？表示一个字符
*/
func PatternMatchs(v string, ps ...string) bool {
	buf := new(bytes.Buffer)
	for _, p := range ps {
		buf.WriteRune('^')
		for _, c := range p {
			switch c {
			case '?':
				buf.WriteRune('.')
			case '*':
				buf.WriteString(".*")
			default:
				buf.WriteRune(c)
			}
		}
		buf.WriteRune('$')
		if ok, _ := regexp.MatchString(buf.String(), v); ok {
			return true
		}
		buf.Reset()
	}
	return false
}

func In(v string, ar []string) bool {
	for _, a := range ar {
		if v == a {
			return true
		}
	}
	return false
}

func DefaultPathGenerator(packageName, serviceName, methodName string) string {
	buf := new(bytes.Buffer)
	buf.WriteRune('/')
	JoinPath(buf, packageName)
	buf.WriteRune('/')
	if strings.HasSuffix(serviceName, ServiceSuffix) {
		JoinPath(buf, serviceName[0:len(serviceName)-ServiceSuffixLength])
	} else {
		JoinPath(buf, serviceName)
	}
	buf.WriteRune('/')
	JoinPath(buf, methodName)
	return buf.String()
}

func JoinPath(buf *bytes.Buffer, path string) {
	for _, ch := range path {
		switch ch {
		case 'A':
			buf.WriteRune('a')
		case 'B':
			buf.WriteRune('b')
		case 'C':
			buf.WriteRune('c')
		case 'D':
			buf.WriteRune('d')
		case 'E':
			buf.WriteRune('e')
		case 'F':
			buf.WriteRune('f')
		case 'G':
			buf.WriteRune('g')
		case 'H':
			buf.WriteRune('h')
		case 'I':
			buf.WriteRune('i')
		case 'J':
			buf.WriteRune('j')
		case 'K':
			buf.WriteRune('k')
		case 'L':
			buf.WriteRune('l')
		case 'M':
			buf.WriteRune('m')
		case 'N':
			buf.WriteRune('n')
		case 'O':
			buf.WriteRune('o')
		case 'P':
			buf.WriteRune('p')
		case 'Q':
			buf.WriteRune('q')
		case 'R':
			buf.WriteRune('r')
		case 'S':
			buf.WriteRune('s')
		case 'T':
			buf.WriteRune('t')
		case 'U':
			buf.WriteRune('u')
		case 'V':
			buf.WriteRune('v')
		case 'W':
			buf.WriteRune('w')
		case 'X':
			buf.WriteRune('x')
		case 'Y':
			buf.WriteRune('y')
		case 'Z':
			buf.WriteRune('z')
		case '.':
			buf.WriteRune('/')
		default:
			buf.WriteRune(ch)
		}
	}
}

func CreateHandlerFunc4Http(tag string, fn func(context.Context, []byte) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			rdata []byte
			wdata []byte
			rsp   interface{}
			err   error
		)
		rdata, err = ioutil.ReadAll(c.Request.Body)
		if err == nil {
			rsp, err = fn(c, rdata)
			if err == nil {
				wdata, _ = json.Marshal(&Response{
					Code: SUCCESS,
					Data: rsp,
					Tag:  tag,
				})
			} else {
				log.Error(c, "%s execute service: %v", tag, err)
				if ersp, ok := err.(*Response); ok {
					wdata, _ = json.Marshal(ersp)
				} else {
					wdata, _ = json.Marshal(&Response{
						Code: EXECUTE_SERVICE_ERROR,
						Msg:  err.Error(),
						Tag:  tag,
					})
				}
			}
		} else {
			log.Error(c, "%s reading request: %v", tag, err)
			wdata, _ = json.Marshal(&Response{
				Code: READING_REQUEST_ERROR,
				Msg:  err.Error(),
				Tag:  tag,
			})
		}
		c.Writer.Header()["Content-Type"] = JsonContentType
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(wdata)
	}
}

func CreateHandlerFunc4Wbsk(tag string, upgrader *websocket.Upgrader, fn func(context.Context, []byte) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error(c, "upgrade connection: %v", tag, err)
			return
		}
		for {
			var (
				mtype int
				rdata []byte
				wdata []byte
				rsp   interface{}
				err   error
			)
			mtype, rdata, err = conn.ReadMessage()
			if err != nil {
				log.Error(c, "%s reading message: %v", tag, err)
				return
			}
			rsp, err = fn(c, rdata)
			if err == nil {
				wdata, _ = json.Marshal(&Response{
					Code: SUCCESS,
					Data: rsp,
					Tag:  tag,
				})
			} else {
				log.Error(c, "%s execute service: %v", tag, err)
				if ersp, ok := err.(*Response); ok {
					wdata, _ = json.Marshal(ersp)
				} else {
					wdata, _ = json.Marshal(&Response{
						Code: EXECUTE_SERVICE_ERROR,
						Msg:  err.Error(),
						Tag:  tag,
					})
				}
			}
			err = conn.WriteMessage(mtype, wdata)
			if err != nil {
				log.Error(c, "%s writing message: %v", tag, err)
				return
			}
		}
	}
}

// 创建upgrader
func CreateWebsocketUpgrader(conf *Config) *websocket.Upgrader {
	upgrader := new(websocket.Upgrader)
	if conf.WbskReadBufferSize != 0 {
		upgrader.ReadBufferSize = conf.WbskReadBufferSize
	}
	if conf.WbskWriteBufferSize != 0 {
		upgrader.WriteBufferSize = conf.WbskWriteBufferSize
	}
	if conf.WbskNotCheckOrigin {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
	return upgrader
}

/*响应结果*/
const (
	SUCCESS               = 0
	UNKNOWN               = -1
	READING_REQUEST_ERROR = 601 // 读取request失败
	PARSING_REQUEST_ERROR = 602 // 解析request失败
	EXECUTE_SERVICE_ERROR = 603 // 执行service失败
)

type Response struct {
	Code int         `json:"code" bson:"code" yaml:"code"`                               // 响应代码
	Msg  string      `json:"msg,omitempty" bson:"msg,omitempty" yaml:"msg,omitempty"`    // 响应消息
	Data interface{} `json:"data,omitempty" bson:"data,omitempty" yaml:"data,omitempty"` // 响应数据
	Tag  string      `json:"tag,omitempty" bson:"tag,omitempty" yaml:"tag,omitempty"`    // 响应标签
}

func (rsp *Response) Error() string {
	bs, _ := json.Marshal(rsp)
	return string(bs)
}

func SuccessResponse(data interface{}, tag ...string) *Response {
	rsp := &Response{
		Data: data,
	}
	if len(tag) > 0 {
		rsp.Tag = tag[0]
	}
	return rsp
}

func FailureResponse(code int, msg string, tag ...string) *Response {
	rsp := &Response{
		Code: code,
		Msg:  msg,
	}
	if len(tag) > 0 {
		rsp.Tag = tag[0]
	}
	return rsp
}

func Error(err error, tag ...string) *Response {
	rsp := &Response{
		Code: UNKNOWN,
		Msg:  err.Error(),
	}
	if len(tag) > 0 {
		rsp.Tag = tag[0]
	}
	return rsp
}

func SplitBase(path string) (string, string) {
	idx := strings.LastIndexByte(path, '/')
	if idx <= 0 {
		return "/", path
	}
	return path[:idx], path[idx:]
}

func Diff(vs1 []string, vs2 []string) (ret []string) {
	if len(vs2) == 0 {
		return vs1
	}
	// 预先生成，避免性能浪费
	ret = make([]string, 0, len(vs1))
__OUTER:
	for _, v1 := range vs1 {
		for _, v2 := range vs2 {
			if v1 == v2 {
				continue __OUTER
			}
		}
		ret = append(ret, v1)
	}
	return
}
