package ctx

import (
	"net/http"
	"strings"

	"github.com/micro-plat/hydra/conf/server"
	"github.com/micro-plat/hydra/conf/server/router"
	"github.com/micro-plat/hydra/context"
	"github.com/micro-plat/hydra/global"
)

var _ context.IPath = &rpath{}

//rpath 处理请求的路径信息
type rpath struct {
	ctx        context.IInnerContext
	serverConf server.IServerConf
}

func newRpath(ctx context.IInnerContext, serverConf server.IServerConf) *rpath {
	return &rpath{
		ctx:        ctx,
		serverConf: serverConf,
	}
}

//GetMethod 获取服务请求方式
func (c *rpath) GetMethod() string {
	return c.ctx.GetMethod()
}

//GetRouter 获取路由信息
func (c *rpath) GetRouter() *router.Router {
	switch c.serverConf.GetMainConf().GetServerType() {
	case global.API, global.Web, global.WS:
		return c.serverConf.GetRouterConf().Match(c.ctx.GetRouterPath(), c.ctx.GetMethod())
	default:
		return router.NewRouter(c.ctx.GetRouterPath(), c.ctx.GetRouterPath(), []string{}, router.WithEncoding("utf-8"))
	}

}

//GetURL 获取请求路径
func (c *rpath) GetURL() string {
	return c.ctx.GetURL().String()
}

//GetRequestPath 获取请求路径
func (c *rpath) GetRequestPath() string {
	return c.ctx.GetURL().Path
}

//GetHeader 获取请求头信息
func (c *rpath) GetHeader(key string) string {
	return strings.Join(c.GetHeaders()[key], ",")
}

//GetHeaders 获取请求的header
func (c *rpath) GetHeaders() http.Header {
	return c.ctx.GetHeaders()
}

//GetHeaders 获取请求的header
func (c *rpath) GetCookies() map[string]string {
	out := make(map[string]string)
	cookies := c.ctx.GetCookies()
	for _, cookie := range cookies {
		out[cookie.Name] = cookie.Value
	}
	return out
}

//GetCookie 获取cookie信息
func (c *rpath) GetCookie(name string) (string, bool) {
	if cookie, ok := c.GetCookies()[name]; ok {
		return cookie, true
	}
	return "", false
}

//GetCookie 获取cookie信息
func (c *rpath) getCookie(name string) string {
	m, _ := c.GetCookie(name)
	return m
}