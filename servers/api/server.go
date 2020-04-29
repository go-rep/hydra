package api

import (
	"context"
	"errors"
	"fmt"
	x "net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/micro-plat/hydra/registry/conf/server/router"
	"github.com/micro-plat/hydra/servers/pkg/middleware"
	"github.com/micro-plat/lib4go/logger"
	"github.com/micro-plat/lib4go/net"
	"github.com/micro-plat/lib4go/types"
)

//Server api服务器
type Server struct {
	name string
	*option
	server  *x.Server
	engine  *gin.Engine
	running bool
	proto   string
	host    string
	port    string
}

//NewServer 创建http api服务嚣
func NewServer(name string, addr string, routers []*router.Router, opts ...Option) (t *Server, err error) {
	t = &Server{
		name: name,
		option: &option{
			serverType:        "api",
			readHeaderTimeout: 6,
			readTimeout:       6,
			writeTimeout:      6,
			metric:            middleware.NewMetric(),
			Logger:            logger.GetSession(name, logger.CreateSession()),
		},
	}
	for _, opt := range opts {
		opt(t.option)
	}
	naddr, err := t.getAddress(addr)
	if err != nil {
		return nil, err
	}
	t.server = &x.Server{
		Addr:              naddr,
		ReadHeaderTimeout: time.Second * time.Duration(t.option.readHeaderTimeout),
		ReadTimeout:       time.Second * time.Duration(t.option.readTimeout),
		WriteTimeout:      time.Second * time.Duration(t.option.writeTimeout),
		MaxHeaderBytes:    1 << 20,
	}
	t.addRouters(routers...)
	return
}

// Run the http server
func (s *Server) Run() error {
	s.running = true
	errChan := make(chan error, 1)
	switch len(s.tls) {
	case 2:
		s.proto = "https"
		go func(ch chan error) {
			if err := s.server.ListenAndServeTLS(s.tls[0], s.tls[1]); err != nil {
				ch <- err
			}
		}(errChan)
	default:
		s.proto = "http"
		go func(ch chan error) {
			if err := s.server.ListenAndServe(); err != nil {
				ch <- err
			}
		}(errChan)

	}
	select {
	case <-time.After(time.Millisecond * 500):
		return nil
	case err := <-errChan:
		s.running = false
		return err
	}
}

//Shutdown 关闭服务器
func (s *Server) Shutdown(timeout time.Duration) {
	if s.server != nil {
		s.metric.Stop()
		s.running = false
		ctx, cannel := context.WithTimeout(context.Background(), timeout)
		defer cannel()
		if err := s.server.Shutdown(ctx); err != nil {
			if err == x.ErrServerClosed {
				s.Infof("%s:已关闭", s.name)
				return
			}
			s.Errorf("关闭出现错误:%v", err)
		}
	}
}

//GetAddress 获取当前服务地址
func (s *Server) GetAddress(h ...string) string {
	if len(h) > 0 && h[0] != "" {
		return fmt.Sprintf("%s://%s:%s", s.proto, h[0], s.port)
	}
	return fmt.Sprintf("%s://%s:%s", s.proto, s.host, s.port)
}

//GetStatus 获取当前服务器状态
func (s *Server) GetStatus() string {
	return types.DecodeString(s.running, true, "运行中", "停止")
}

func (s *Server) getAddress(addr string) (string, error) {
	host := "0.0.0.0"
	port := "8081"
	args := strings.Split(addr, ":")
	l := len(args)
	if addr == "" {
		l = 0
	}
	switch l {
	case 0:
	case 1:
		if govalidator.IsPort(args[0]) {
			port = args[0]
		} else {
			host = args[0]
		}
	case 2:
		host = args[0]
		port = args[1]
	default:
		return "", fmt.Errorf("%s地址不合法", addr)
	}
	switch host {
	case "0.0.0.0", "":
		s.host = net.GetLocalIPAddress()
	case "127.0.0.1", "localhost":
		s.host = host
	default:
		if net.GetLocalIPAddress(host) != host {
			return "", fmt.Errorf("%s地址不合法", addr)
		}
		s.host = host
	}

	if !govalidator.IsPort(port) {
		return "", fmt.Errorf("%s端口不合法", addr)
	}
	if port == "80" {
		if err := checkPrivileges(); err != nil {
			return "", err
		}
	}
	s.port = port
	return fmt.Sprintf("%s:%s", host, s.port), nil
}
func checkPrivileges() error {
	if output, err := exec.Command("id", "-g").Output(); err == nil {
		if gid, parseErr := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 32); parseErr == nil {
			if gid == 0 {
				return nil
			}
			return errRootPrivileges
		}
	}
	return errUnsupportedSystem
}

var errUnsupportedSystem = errors.New("Unsupported system")
var errRootPrivileges = errors.New("You must have root user privileges. Possibly using 'sudo' command should help")
