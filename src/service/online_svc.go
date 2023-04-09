package service

import (
	"github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
	"interview/common/resource"
	"interview/conf"
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

type IssueStruct struct {
	Id int `json:"id"`
	//	Score int `json:"score"`
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
