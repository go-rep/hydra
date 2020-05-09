package remove

import (
	"github.com/micro-plat/cli/cmds"
	"github.com/micro-plat/hydra/application"
	"github.com/micro-plat/hydra/hydra/cmds/daemon"
	"github.com/micro-plat/lib4go/errs"
	"github.com/urfave/cli"
)

func init() {
	cmds.Register(
		cli.Command{
			Name:   "remove",
			Usage:  "删除服务",
			Action: doRemove,
		})
}
func doRemove(c *cli.Context) (err error) {

	//关闭日志显示
	application.Current().Log().Pause()
	service, err := daemon.New(application.AppName, application.AppName)
	if err != nil {
		return err
	}
	msg, err := service.Remove()
	if err != nil {
		return err
	}
	return errs.NewIgnoreError(0, msg)
}