package main

import (
	"flag"
	"github.com/golang/glog"
	cConfig "interview/common/config"
	"interview/conf"
	"interview/src/model"
	"interview/src/service"
	"net/http"
)

// TODO 看下有没能用上kafka的地方，改良
// TODO 根据路由选择不同的方法去执行
// TODO 写测试&性能测试
// TODO 优雅error处理
func main() {
	//接收参数，第一版可以接收命令行参数
	flag.Parse()
	//主要逻辑
	http.HandleFunc("/", service.HomeHandler)
	http.HandleFunc("/reset", service.ResetHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

var (
	flagConfDir string //配置文件路径
)

// 初始化各种需要用到的组件
func init() {
	flag.StringVar(&flagConfDir, "conf", "/Applications/go/interview/conf", "config file dir")
	//初始化配置
	initConfig()
	//初始化mysql
	model.InitSqlModel()
	//初始化redis
	service.Init()
}

//TODO 升级：根据传入的类型和优先级，组成一个唯一key，作为redis的分组，判断该组是否存在，存在则继续从该组获取数据；
//这个升级会有不兼容：mysql的isread字段无法区分对应问题应该属于哪个分组

// 初始化配置文件
func initConfig() {
	// 初始化配置文件
	config := conf.GetIns()
	configManager := cConfig.NewConfigManager(flagConfDir+"/base.json", config)
	ret := configManager.Load()
	if !ret {
		glog.Fatal("初始化配置文件失败")
		return
	}
	// configManager.StartAutoLoad(flags.FlagConfUpdateInterval)
	glog.V(1).Infoln("服务配置:", config)
}
