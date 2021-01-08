package ws

import "strings"

type option struct {
	Status    string `json:"status,omitempty" valid:"in(start|stop)" toml:"status,omitempty"`
	RTimeout  int    `json:"rTimeout,omitempty" toml:"rTimeout,omitzero"`
	WTimeout  int    `json:"wTimeout,omitempty" toml:"wTimeout,omitzero"`
	RHTimeout int    `json:"rhTimeout,omitempty" toml:"rhTimeout,omitzero"`
	Host      string `json:"host,omitempty" toml:"host,omitempty"`
	Domain    string `json:"dns,omitempty" toml:"dns,omitempty"`
	Trace     bool   `json:"trace,omitempty" toml:"trace,omitempty"`
}

//Option 配置选项
type Option func(*option)

//WithTrace 构建api server配置信息
func WithTrace() Option {
	return func(a *option) {
		a.Trace = true
	}
}

//WithTimeout 构建api server配置信息
func WithTimeout(rtimeout int, wtimout int) Option {
	return func(a *option) {
		a.RTimeout = rtimeout
		a.WTimeout = wtimout
	}
}

//WithHeaderReadTimeout 构建api server配置信息
func WithHeaderReadTimeout(htimeout int) Option {
	return func(a *option) {
		a.RHTimeout = htimeout
	}
}

//WithHost 设置host
func WithHost(host ...string) Option {
	return func(a *option) {
		a.Host = strings.Join(host, ";")
	}
}

//WithDisable 禁用任务
func WithDisable() Option {
	return func(a *option) {
		a.Status = StartStop
	}
}

//WithEnable 启用任务
func WithEnable() Option {
	return func(a *option) {
		a.Status = StartStatus
	}
}

//WithDNS 设置请求域名
func WithDNS(host string, ip ...string) Option {
	return func(a *option) {
		a.Host = host
		if len(ip) > 0 {
			a.Domain = ip[0]
		}
	}
}
