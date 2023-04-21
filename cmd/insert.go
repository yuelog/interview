package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/glog"
	cConfig "interview/common/config"
	"interview/conf"
	"interview/src/model"
	"strings"
)

var issueStruct model.IssueStruct

func main() {
	flag.Parse()
	loadData()
	err := model.InsertData(&issueStruct, issueType, priority)
	if err != nil {
		panic(err)
	}
	fmt.Println("insert success")
}

var (
	flagConfDir  string //配置文件路径
	issue        string
	tips         string
	answer       string
	relatedIssue string
	knowledge    string
	issueType    string
	priority     int
)

func loadData() {
	issueStruct.Issue = issue
	issueStruct.Answer = answer
	issueStruct.RelatedIssues = stringToJson(relatedIssue)
	issueStruct.Knowledge = stringToJson(knowledge)
}

func init() {
	flag.StringVar(&flagConfDir, "conf", "../conf", "config file dir")
	flag.StringVar(&issue, "issue", "", "issue content")
	flag.StringVar(&tips, "tips", "", "tips content")
	flag.StringVar(&answer, "answer", "", "answer content")
	flag.StringVar(&relatedIssue, "related_issue", "", "related issue content")
	flag.StringVar(&knowledge, "knowledge", "", "knowledge content")
	flag.StringVar(&issueType, "type", "", "type content")
	flag.IntVar(&priority, "priority", 1, "priority number")

	//初始化配置
	initConfig()
	//初始化mysql
	model.InitSqlModel()
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

func stringToJson(str string) string {
	if str == "" {
		return str
	}
	slice := strings.Split(str, ",")
	// 将切片转换为JSON格式
	jsonBytes, err := json.Marshal(slice)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}
