package service

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
	"interview/common/resource"
	"interview/conf"
	"interview/src/model"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

const MY_REDIS_NAME = "online_redis"

func Init() {
	config := conf.GetIns()

	glog.V(1).Infoln("初始化redis,配置为:", config)
	redisConfig := &resource.RedisPoolConfig{
		Address:  config.OnlineHost,
		Password: config.OnlinePassword,
		DbIndex:  config.OnlineDatabase,
		// 空闲连接超时时间，超过超时时间的空闲连接会被关闭。
		// 如果设置成0，空闲连接将不会被关闭
		// 应该设置一个比redis服务端超时时间更短的时间
		IdleTimeout: config.OnlinePoolIdleTimeout,
		//最大空闲连接数
		MaxIdle: config.OnlinePoolMaxIdle,
		// 一个pool所能分配的最大的连接数目
		// 当设置成0的时候，该pool连接数没有限制
		MaxActive: config.OnlinePoolMaxActive,
		// 如果Wait被设置成true，则Get()方法将会阻塞
		Wait:           config.OnlinePoolWait,
		ConnectTimeout: config.OnlineConnectTimeout,
		ReadTimeout:    config.OnlineReadTimeout,
		WriteTimeout:   config.OnlineWriteTimeout,
	}
	resource.RedisSetup(MY_REDIS_NAME, redisConfig)
}

// 获取Redis中的面试题
func GetIssueId(key string) (int, bool) {
	conn, err := resource.GetRedisConnect(MY_REDIS_NAME)
	if err != nil {
		glog.Errorln("redis连接失败")
		return 0, false
	}
	defer conn.Close()
	//没有数据,则需要加载到redis
	id, err := redis.Int(conn.Do("LPOP", key))
	if err != nil {
		//每次新加载到redis会触发一次这个err，但是不影响正常逻辑
		//glog.Errorln("redis获取zset数据失败, err:", err)
		return 0, false
	}

	return id, true
}

func LoadDatasToRedis(issueIds []int, key string) {
	conn, err := resource.GetRedisConnect(MY_REDIS_NAME)
	if err != nil {
		glog.Errorln("redis连接失败")
		return
	}
	defer conn.Close()

	args := []interface{}{key}
	for _, id := range issueIds {
		args = append(args, id)
	}
	_, err = conn.Do("LPUSH", args...)
	if err != nil {
		glog.Errorln("redis写入失败, err:", err)
	}
	return
}

func ClearRedis(key string) {
	conn, err := resource.GetRedisConnect(MY_REDIS_NAME)
	if err != nil {
		glog.Errorln("redis连接失败")
		return
	}
	_, err = conn.Do("DEL", key)
	if err != nil {
		glog.Errorln("redis删除失败, err:", err)
	}
	return
}

// 输入:类型、优先级
// 输出：满足条件 且 未读 的题目一条
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.RequestURI() == "/favicon.ico" {
		return
	}
	pri := r.URL.Query().Get("pri")
	issueType := r.URL.Query().Get("type")
	if pri == "" {
		pri = "1"
	}
	key := genKey(pri, issueType)
	priority, err := strconv.Atoi(pri)
	if err != nil {
		glog.Errorln("字符串转换为整数时出错：", err)
		return
	}
	//从redis中获取数据，redis使用zset存储，只存id => score
	id, succ := GetIssueId(key)
	//没有数据，从mysql载入
	if !succ {
	Retry:
		//根据类型、优先级获得未读过的题目加载进redis（这样就考虑了redis数据过期的问题，只要不调用reset接口，就可以从上次的进度继续
		issueIds := model.GetIssueIds(priority, issueType)
		if len(issueIds) == 0 {
			fmt.Fprintf(w, "<h1>题库已空，重新测验请先执行reset</h1>")
			return
		}
		//实现"随机":打乱取出的id，分配score
		rand.Shuffle(len(issueIds), func(i, j int) {
			issueIds[i], issueIds[j] = issueIds[j], issueIds[i]
		})
		LoadDatasToRedis(issueIds, key)
		//从redis中取出一条id，到mysql取出对应数据并更新mysql对应数据为已读
		//之所以使用Redis还要更新MySQL的状态，是为了保证一致性（比如隔一天但不想重置数据，若redis数据过期，此时只要将MySQL中未读数据装载进Redis即可）
		id, succ = GetIssueId(key)
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
	//fmt.Println(data)
	fmt.Fprintf(w, formatResult(data))
}

// Reset 重置:把所有数据都置为未读状态,并把数据加载到redis
// TODO 数据加载到Redis这步可以先做一部分，后面升级为把所有类型的都加载到redis
func ResetHandler(w http.ResponseWriter, r *http.Request) {
	//把所有数据都置为未读状态
	model.Reset()
	//把数据加载到Redis
	priority := 1
	issueType := ""
	issueIds := model.GetIssueIds(priority, issueType)
	//实现"随机":打乱取出的id，分配score
	rand.Shuffle(len(issueIds), func(i, j int) {
		issueIds[i], issueIds[j] = issueIds[j], issueIds[i]
	})
	key := genKey(strconv.Itoa(priority), issueType)
	ClearRedis(key)
	LoadDatasToRedis(issueIds, key)
	fmt.Fprint(w, "<html><body><h1>Success!</h1></body></html>")
}

func genKey(priority string, issueType string) string {
	return priority + "_" + issueType
}

func formatResult(data *model.IssueStruct) string {
	start := "<html><body><h1>" + data.Issue + "</h1><br/>"
	end := "</body></html>"
	body := ""
	body = formatAnswer(body, data.Answer, "答案")
	body = formatBody(body, data.RelatedIssues, "相关问题")
	body = formatBody(body, data.Knowledge, "相关知识点")

	return start + body + end
}

func formatAnswer(body, datas, fieldName string) string {
	if datas == "" {
		return body
	}
	dataSlice := strings.Split(datas, ",")
	for i, data := range dataSlice {
		body += "<a href=" + data + " target=\"_blank\"> " + fieldName + strconv.Itoa(i+1) + "</a>&nbsp;&nbsp;&nbsp;&nbsp;"
	}
	body += "<br/><br/>"
	return body
}

func formatBody(body, jsonData, fieldName string) string {
	if jsonData == "" {
		return body
	}
	var datas []string
	err := json.Unmarshal([]byte(jsonData), &datas)
	if err != nil {
		glog.Errorln("Error:", err)
		return ""
	}
	body += "<h2>" + fieldName + "</h2>"
	for i, data := range datas {
		body += "<p style='color:" + randColor(i) + "'>" + data + "</font>"
	}
	body += "<br/><br/>"
	return body
}

func randColor(num int) string {
	colorMap := map[int]string{0: "black", 1: "blue", 2: "red"}
	return colorMap[(num+1)%len(colorMap)]
}
