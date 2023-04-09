package resource

import (
	"errors"
	"github.com/golang/glog"
	"github.com/gomodule/redigo/redis"
	"time"
)

// redis pool的配置
type RedisPoolConfig struct {
	Address        string //redis host
	Password       string //redis 密码
	DbIndex        int    //数据库index
	IdleTimeout    int    //空闲连接的超时时间，为0空闲连接将不会被关闭（应该要设置成比redis客户端超时时间更短的时间）
	MaxIdle        int    //最大空闲连接数
	MaxActive      int    //连接池的最大连接数，为0则无限制
	Wait           bool   //Get()获取连接实例的方法会不会阻塞，如果为true则阻塞等待到能获取到一个连接实例位置，如果为false则会直接返回nil
	ConnectTimeout int    //redis连接超时
	ReadTimeout    int    //redis读取操作
	WriteTimeout   int    //redis写入超时
}

var (
	rPool map[string]*redis.Pool
)

func init() {
	if len(rPool) <= 0 {
		rPool = make(map[string]*redis.Pool)
	}
}

// 简单的生成默认配置
func SimpleConfig(address string, password string) *RedisPoolConfig {
	return &RedisPoolConfig{
		Address:        address,
		Password:       password,
		DbIndex:        0,
		IdleTimeout:    3,
		MaxIdle:        100,  //最大空闲连接数
		MaxActive:      2048, //最大连接数
		Wait:           true, //要阻塞，不阻塞业务逻辑可能就失败了
		ConnectTimeout: 3,
		ReadTimeout:    3,
		WriteTimeout:   3,
	}
}

/**
 * 初始化 某个服务的redis连接
 */
//初始化redis连接后，装载进rPool中
func RedisSetup(name string, config *RedisPoolConfig) {
	address := config.Address
	password := config.Password
	dbIndex := config.DbIndex
	idleTimeOut := config.IdleTimeout
	maxIdle := config.MaxIdle
	maxActive := config.MaxActive
	wait := config.Wait
	serverPool := &redis.Pool{
		//连接redis的方法
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp",
				address, redis.DialPassword(password),
				redis.DialConnectTimeout(time.Duration(config.ConnectTimeout)*time.Second),
				redis.DialReadTimeout(time.Duration(config.ReadTimeout)*time.Second),
				redis.DialWriteTimeout(time.Duration(config.WriteTimeout)*time.Second))
			if err != nil {
				glog.Errorf("[redis_pool] - 创建连接实例失败,name:%v,config:%+v,err:%v \n", name, config, err.Error())
				return nil, err
			}
			if dbIndex != 0 { //如果指定了db，则切换到对应的db
				c.Do("SELECT", dbIndex)
			}
			glog.V(3).Infof("[redis_pool] - 创建连接实例成功,name:%v,address:%v \n", name, address)
			return c, nil
		},
		//使用前先测试是否可用
		//放回连接池小于1分钟的连接直接返回，否则ping一下看是否正常，正常就返回
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
		//空闲连接保持3s后关闭
		IdleTimeout: time.Duration(idleTimeOut) * time.Second,
		MaxIdle:     maxIdle,   //最大空闲连接数 100
		MaxActive:   maxActive, // 一个pool所能分配的最大的连接数目 2048
		Wait:        wait}      // 如果Wait被设置成true，则Get()方法将会阻塞  //要阻塞，不阻塞业务逻辑可能就失败了

	rPool[name] = serverPool
}

/**
 * 获取某个服务的redis连接
 */
func GetRedisConnect(name string) (redis.Conn, error) {
	if conn, ok := rPool[name]; ok {
		return conn.Get(), nil
	}
	return nil, errors.New("[redis_pool] - redis还未初始化,必须先执行RedisSetup(),name:" + name)
}
