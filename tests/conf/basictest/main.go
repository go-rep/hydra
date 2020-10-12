package main

import (
	"github.com/micro-plat/hydra"
	_ "github.com/micro-plat/hydra/components/pkgs/mq/redis"
	"github.com/micro-plat/hydra/conf/server/api"
	"github.com/micro-plat/hydra/conf/server/apm"
	"github.com/micro-plat/hydra/conf/server/auth/basic"
	"github.com/micro-plat/hydra/context"
	"github.com/micro-plat/hydra/hydra/servers/http"
)

/*
	1.测试通过;
*/

var app = hydra.NewApp(
	hydra.WithServerTypes(http.API),
	hydra.WithPlatName("taosy-basic-test"),
	hydra.WithSystemName("test-basic"),
	hydra.WithClusterName("taosybasic"),
	hydra.WithRegistry("redis://192.168.5.79:6379"),
	// hydra.WithRegistry("redis://192.168.0.111:6379,192.168.0.112:6379,192.168.0.113:6379,192.168.0.114:6379,192.168.0.115:6379,192.168.0.116:6379"),
	// hydra.WithRegistry("zk://192.168.0.101:2181"),
)

func init() {
	app.OnHandleExecuting(func(ctx context.IContext) (r interface{}) {
		ctx.Log().Info("执行前处理")
		ctx.Log().Info("执行路径:", ctx.Request().Path().GetRouter().Service)

		return
	})

	serverApi := hydra.Conf.API(":8070", api.WithTimeout(10, 10), api.WithDisable(), api.WithTrace(), api.WithHost("192.168.5.107"), api.WithDNS(""))
	serverApi.APM("skywalking", apm.WithDisable())
	serverApi.Basic(basic.WithUP("taosy", "123456"), basic.WithUP("taosy1", "12345"), basic.WithExcludes("/taosy/testapi/:name"))
	app.API("/taosy/testapi/:name", func(ctx context.IContext) (r interface{}) {
		ctx.Log().Info("api 接口服务测试")
		ctx.Log().Info("name :", ctx.Request().Param("name"))
		return "success"
	})

	app.API("/taosy/testapi", func(ctx context.IContext) (r interface{}) {
		ctx.Log().Info("api 接口服务测试")
		ctx.Log().Info("name :", ctx.Request().Param("name"))

		return "success"
	})

	app.API("/taosy/jwttest", func(ctx context.IContext) (r interface{}) {
		ctx.Log().Info("jwt 接口服务测试")

		return "success"
	})

}

func main() {
	app.Start()
}