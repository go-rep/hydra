package cmds

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/micro-plat/hydra"
	"github.com/micro-plat/hydra/hydra/servers/http"
	"github.com/micro-plat/hydra/test/assert"
)

func Test_stop_Normal_running(t *testing.T) {
	resetServiceName(t.Name())
	execPrint(t)
	//正常的开启
	fileName := fmt.Sprint(time.Now().Nanosecond())
	file, _ := os.Create(fileName)
	orgStd := *os.Stdout
	*os.Stdout = *file
	//defer os.Remove(fileName)

 	var app = hydra.NewApp(
		hydra.WithServerTypes(http.API),
		hydra.WithPlatName("xxtest"),
		hydra.WithSystemName("apiserver"),
		hydra.WithClusterName("c"),
	)
	hydra.Conf.API(":19020")

	//1. 删除服务
 	os.Args = []string{"xxtest", "remove"}
	app.Start()
	time.Sleep(time.Second * 2)

	//2. 安装服务
	os.Args = []string{"xxtest", "install", "-r", runRegistryAddr, "-c", "c"}
	go app.Start()
	time.Sleep(time.Second * 2)

	//3. 启动服务
 	os.Args = []string{"xxtest", "start"}
	go app.Start()

	time.Sleep(time.Second)

	//4. 关闭服务状态
 	os.Args =  []string{"xxtest", "stop"}
	go app.Start()
	time.Sleep(time.Second)

	//5. 清除服务
 	os.Args = []string{"xxtest", "remove"}
	go app.Start()

	//还原std
	*os.Stdout = orgStd

	file.Close()
	time.Sleep(time.Second)
	bytes, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Error(err)
		return
	}
	line := string(bytes)
	lines := strings.Split(line,"\r")
	for _,row := range lines{
		if strings.Contains(row,"Stopping"){
			line = row
			break
		}		
	}
 
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" { 
		result := strings.Contains(line, "sudo") || (strings.Contains(line, "Stopping") && strings.Contains(line, "OK"))
		assert.Equal(t, true, result, "关闭正常服务运行")
	}
	if runtime.GOOS == "windows" {
		result :=(strings.Contains(line, "Stopping") && strings.Contains(line, "OK"))
		assert.Equal(t, true, result, "关闭正常服务运行")
	} 
	time.Sleep(time.Second)
}

func Test_stop_Not_installed(t *testing.T) {
	resetServiceName(t.Name())
	//启动未安装的服务
	execPrint(t)

	fileName := fmt.Sprint(time.Now().Nanosecond())
	file, _ := os.Create(fileName)
	orgStd := *os.Stdout
	*os.Stdout = *file
	defer os.Remove(fileName)
 
	var app = hydra.NewApp(
		hydra.WithServerTypes(http.API),
		hydra.WithPlatName("xxtest"),
		hydra.WithSystemName("apiserver"),
		hydra.WithClusterName("c"),
	)
	hydra.Conf.API(":19010")

	//1. 清除服务(保证没有服务安装)
	os.Args = []string{"xxtest", "remove"}
	go app.Start()
	time.Sleep(time.Second * 2)

	//2. 启动服务
	os.Args = []string{"xxtest", "stop"}
	go app.Start()
	time.Sleep(time.Second * 2)

	//还原std
	*os.Stdout = orgStd

	file.Close()
	time.Sleep(time.Second)
	bytes, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Error(err)
		return
	}
	line := string(bytes)

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		//unbuntu/centos
		result := strings.Contains(line, "sudo") || strings.Contains(line, "Service is not installed")
		assert.Equal(t, true, result, "启动未安装的服务")
	}
	if runtime.GOOS == "windows" {
		result := strings.Contains(line, "not exist as an installed service")
		assert.Equal(t, true, result, "启动未安装的服务")
	}

	time.Sleep(time.Second)
}


func Test_stop_has_stopped(t *testing.T) {
	resetServiceName(t.Name())
	//启动未安装的服务
	execPrint(t)

	fileName := fmt.Sprint(time.Now().Nanosecond())
	file, _ := os.Create(fileName)
	orgStd := *os.Stdout
	*os.Stdout = *file
//	defer os.Remove(fileName)


	var app = hydra.NewApp(
		hydra.WithServerTypes(http.API),
		hydra.WithPlatName("xxtest"),
		hydra.WithSystemName("apiserver"),
		hydra.WithClusterName("c"),
	)
	hydra.Conf.API(":19010")

	//1. 清除服务(保证没有服务安装)
	os.Args = []string{"xxtest", "remove"}
	go app.Start()
	time.Sleep(time.Second * 2)

	//2. 安装服务
	os.Args = []string{"xxtest", "install", "-r", runRegistryAddr, "-c", "c"}
	go app.Start()
	time.Sleep(time.Second * 2)

	
	//3. 关闭服务状态
	os.Args =  []string{"xxtest", "stop"}
	go app.Start()
	time.Sleep(time.Second)

	//还原std
	*os.Stdout = orgStd

	file.Close()
	time.Sleep(time.Second)
	bytes, err := ioutil.ReadFile(fileName)

	if err != nil {
		t.Error(err)
		return
	}
	line := string(bytes)

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		//unbuntu/centos
		result := strings.Contains(line, "sudo") || strings.Contains(line, "has already been stopped")
		assert.Equal(t, true, result, "启动未安装的服务")
	}
	if runtime.GOOS == "windows" {
		result := strings.Contains(line, "not exist as an installed service")
		assert.Equal(t, true, result, "启动未安装的服务")
	}

	time.Sleep(time.Second)
}
