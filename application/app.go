package application

import (
	"fmt"
	"strings"

	"github.com/micro-plat/hydra/context"
	"github.com/micro-plat/hydra/registry/conf/server"
	"github.com/micro-plat/lib4go/logger"
	"github.com/micro-plat/lib4go/types"
)

var traces = []string{"cpu", "mem", "block", "mutex", "web"}

//DefApp 默认app
var DefApp = &application{
	log:   logger.New("hydra"),
	close: make(chan struct{}),
}

type application struct {

	//registryAddr 集群地址
	RegistryAddr string `json:"registryAddr" valid:"ascii,required"`

	//PlatName 平台名称
	PlatName string `json:"platName" valid:"ascii,required"`

	//SysName 系统名称
	SysName string `json:"sysName" valid:"ascii,required"`

	//ServerTypes 服务器类型
	ServerTypes []string `json:"serverTypes" valid:"in(api|web|rpc|ws|mqc|cron),required"`

	//ServerTypeNames 服务类型名称
	ServerTypeNames string

	//ClusterName 集群名称
	ClusterName string `json:"clusterName" valid:"ascii,required"`

	//Name 服务器请求名称
	Name string

	//Trace 显示请求与响应信息
	Trace string `valid:"in(cpu|mem|block|mutex|web)"`

	//isClose 是否关闭当前应用程序
	isClose bool

	//log 日志管理
	log logger.ILogger

	//close 关闭通道
	close chan struct{}
}

func (m *application) Bind() (err error) {
	//处理参数
	if err := m.check(); err != nil {
		return err
	}

	//增加调试参数
	if IsDebug {
		m.PlatName += "_debug"
	}
	return nil
}

//Server 获取服务器配置信息
func (m *application) Server(tp string) server.IServerConf {
	s, err := server.Cache.GetServerConf(tp)
	if err == nil {
		return s
	}
	panic(fmt.Errorf("[%s]服务器未启动:%w", tp, err))
}

//CurrentContext 获取当前请求上下文
func (m *application) CurrentContext() context.IContext {
	return nil
}

//GetRegistryAddr 注册中心
func (m *application) GetRegistryAddr() string {
	return m.RegistryAddr
}

//GetPlatName 平台名称
func (m *application) GetPlatName() string {
	return m.PlatName
}

//GetSysName 系统名称
func (m *application) GetSysName() string {
	return m.SysName
}

//GetServerTypes 服务器类型
func (m *application) GetServerTypes() []string {
	return m.ServerTypes
}

//GetClusterName 集群名称
func (m *application) GetClusterName() string {
	return m.ClusterName
}

//GetTrace 显示请求与响应信息
func (m *application) GetTrace() string {
	return m.Trace
}

//ClosingNotify 获取系统关闭通知
func (m *application) ClosingNotify() chan struct{} {
	return m.close
}

//Log 获取日志组件
func (m *application) Log() logger.ILogger {
	return m.log
}

//Close 显示请求与响应信息
func (m *application) Close() {
	m.isClose = true
	close(m.close)
}
func parsePath(p string) (platName string, systemName string, serverTypes []string, clusterName string, err error) {
	fs := strings.Split(strings.Trim(p, "/"), "/")
	if len(fs) != 4 {
		err := fmt.Errorf("系统名称错误，格式:/[platName]/[sysName]/[typeName]/[clusterName]")
		return "", "", nil, "", err
	}
	serverTypes = strings.Split(fs[2], "-")
	platName = fs[0]
	systemName = fs[1]
	clusterName = fs[3]
	return
}
func (m *application) check() (err error) {

	if m.ServerTypeNames != "" {
		m.ServerTypes = strings.Split(m.ServerTypeNames, "-")
	}
	if m.Name != "" {
		m.PlatName, m.SysName, m.ServerTypes, m.ClusterName, err = parsePath(m.Name)
		if err != nil {
			return err
		}
	}

	if m.RegistryAddr == "" {
		return fmt.Errorf("注册中心地址不能为空")
	}
	if m.PlatName == "" {
		return fmt.Errorf("平台名称不能为空")
	}
	if m.SysName == "" {
		return fmt.Errorf("系统名称不能为空")
	}
	if len(m.ServerTypes) == 0 {
		return fmt.Errorf("服务器类型不能为空")
	}
	if m.ClusterName == "" {
		return fmt.Errorf("集群名称不能为空")
	}
	if m.Trace != "" && !types.StringContains(traces, m.Trace) {
		return fmt.Errorf("trace名称只能是%v", traces)
	}

	return nil
}