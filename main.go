package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	cConfig "interview/common/config"
	"interview/conf"
	"interview/src/model"
	"interview/src/service"
	"math/rand"
	"strconv"
)

// TODO 看下有没能用上kafka的地方，改良
// TODO 根据路由选择不同的方法去执行
// TODO 写测试&性能测试
func main() {
	//接收参数，第一版可以接收命令行参数
	flag.Parse()
	//TODO 后面修改接收web传参（兼容）

	//主要逻辑
	Do()
}

var (
	flagConfDir  string //配置文件路径
	flagType     string // 问题类型
	flagPriority int    //优先级
)

// 初始化各种需要用到的组件
func init() {
	flag.StringVar(&flagConfDir, "conf", "/Applications/go/interview/conf", "config file dir")
	flag.StringVar(&flagType, "type", "", "issue type")
	flag.IntVar(&flagPriority, "pri", 1, "priority")
	//初始化配置
	initConfig()
	//初始化mysql
	model.InitSqlModel()
	//初始化redis
	service.Init()
}

//TODO 升级：根据传入的类型和优先级，组成一个唯一key，作为redis的分组，判断该组是否存在，存在则继续从该组获取数据；
//这个升级会有不兼容：mysql的isread字段无法区分对应问题应该属于哪个分组

// 输入:类型、优先级
// 输出：满足条件 且 未读 的题目一条
func Do() {
	key := strconv.Itoa(flagPriority) + "_" + flagType
	//从redis中获取数据，redis使用zset存储，只存id => score
	id, succ := service.GetIssueId(key)
	//没有数据，从mysql载入
	if !succ {
	Retry:
		//根据类型、优先级获得未读过的题目加载进redis（这样就考虑了redis数据过期的问题，只要不调用reset接口，就可以从上次的进度继续
		issueIds := model.GetIssue(flagPriority, flagType)
		if issueIds == nil {
			goto Retry
		}
		//实现"随机":打乱取出的id，分配score
		rand.Shuffle(len(issueIds), func(i, j int) {
			issueIds[i], issueIds[j] = issueIds[j], issueIds[i]
		})
		service.LoadDatasToRedis(issueIds, key)
		//从redis中取出一条id，到mysql取出对应数据并更新mysql对应数据为已读(用事务)
		//之所以使用Redis还要更新MySQL的状态，是为了保证一致性（比如隔一天但不想重置数据，此时只要将MySQL中未读数据装载进Redis即可）
		id, succ = service.GetIssueId(key)
		if !succ {
			goto Retry
		}
	}
	//取数据并更新为已读
	data, err := model.GetIssueById(id)
	if err != nil {
		panic(err)
	}
	//一些数据拼接，比如answer拼接到html标签中
	fmt.Println(data)
}

// Reset 重置:把所有数据都置为未读状态
func Reset() {

}

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
