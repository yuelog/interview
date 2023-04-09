package config

// 支持自动更新的Config
import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	io "io/ioutil"
	"time"
)

type ConfigManager struct {
	ts             int64           // 上次更新时间
	filename       string          // 文件路径
	configPoint    ConfigInterface // 配置指针
	updateInterval int64
}

type ConfigInterface interface {
	CheckDefault()
}

func NewConfigManager(filename string, config ConfigInterface) *ConfigManager {
	return &ConfigManager{filename: filename, ts: 0, configPoint: config, updateInterval: 0}
}

func (manager *ConfigManager) Load() bool {
	now := time.Now().Unix()

	if (now - manager.ts) < manager.updateInterval {
		glog.Warningf("load file too often: %d", now-manager.ts)
		return false
	}

	data, err := io.ReadFile(manager.filename)
	if err != nil {
		fmt.Println(err)
		glog.Errorf("load file fail: %s", manager.filename)
		return false
	}

	bytes := []byte(data)

	err = json.Unmarshal(bytes, manager.configPoint)

	if err != nil {
		glog.Errorf("load file fail:%s, error:%s", manager.filename, err.Error())
		return false
	}

	//检查，若没有赋值就赋默认值
	manager.configPoint.CheckDefault()
	manager.ts = now
	return true
}

func (manager *ConfigManager) StartAutoLoad(updateInterval int) {
	manager.updateInterval = int64(updateInterval)
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			ret := manager.Load()
			if !ret {
				continue
			}
			manager.configPoint.CheckDefault()
			glog.V(5).Infof("update config succ %+v", manager.configPoint)
		}
	}()
}
