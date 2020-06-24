package static

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/asaskevich/govalidator"
	"github.com/micro-plat/hydra/conf"
	"github.com/micro-plat/hydra/global"
	"github.com/micro-plat/lib4go/archiver"
)

//IStatic 静态文件接口
type IStatic interface {
	GetConf() (*Static, bool)
}

//Static 设置静态文件配置
type Static struct {
	Dir       string   `json:"dir,omitempty" valid:"ascii" toml:"dir,omitempty"`
	Archive   string   `json:"archive,omitempty" valid:"ascii" toml:"archive,omitempty"`
	Prefix    string   `json:"prefix,omitempty" valid:"ascii" toml:"prefix,omitempty"`
	Exts      []string `json:"exts,omitempty" valid:"ascii" toml:"exts,omitempty"`
	Exclude   []string `json:"exclude,omitempty" valid:"ascii" toml:"exclude,omitempty"`
	FirstPage string   `json:"first-page,omitempty" valid:"ascii" toml:"first-page,omitempty"`
	Rewriters []string `json:"rewriters,omitempty" valid:"ascii" toml:"rewriters,omitempty"`
	Disable   bool     `json:"disable,omitempty" toml:"disable,omitempty"`
}

//New 构建静态文件配置信息
func New(opts ...Option) *Static {
	s := newStatic()
	for _, opt := range opts {
		opt(s)
	}
	return s
}

//AllowRequest 是否是合适的请求
func (s *Static) AllowRequest(m string) bool {
	return m == http.MethodGet || m == http.MethodHead
}

type ConfHandler func(cnf conf.IMainConf) *Static

func (h ConfHandler) Handle(cnf conf.IMainConf) interface{} {
	return h(cnf)
}

//GetConf 设置static
func GetConf(cnf conf.IMainConf) *Static {
	//设置静态文件路由
	static := Static{}
	_, err := cnf.GetSubObject("static", &static)
	if err != nil && err != conf.ErrNoSetting {
		panic(fmt.Errorf("static配置有误:%v", err))
	}
	if err == conf.ErrNoSetting {
		static.Disable = true
		return &static
	}
	if b, err := govalidator.ValidateStruct(&static); !b {
		panic(fmt.Errorf("static配置有误:%v", err))
	}
	static.Dir, err = unarchive(static.Dir, static.Archive) //处理归档文件
	return &static
}

var waitRemoveDir = make([]string, 0, 1)

func unarchive(dir string, path string) (string, error) {
	if path == "" {
		return dir, nil
	}
	reader, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return dir, nil
		}
		return "", fmt.Errorf("无法打开文件:%w", err)
	}
	archive := archiver.MatchingFormat(path)
	if archive == nil {
		return "", fmt.Errorf("指定的文件不是归档文件:%s", path)
	}
	tmpDir, err := ioutil.TempDir("", "hydra")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败:%v", err)
	}

	defer reader.Close()
	ndir := filepath.Join(tmpDir, dir)
	err = archive.Read(reader, ndir)
	if err != nil {
		return "", fmt.Errorf("读取归档文件失败:%v", err)
	}
	waitRemoveDir = append(waitRemoveDir, tmpDir)
	return ndir, nil
}
func init() {
	global.Def.AddCloser(func() error {
		for _, d := range waitRemoveDir {
			os.RemoveAll(d)
		}
		return nil
	})
}