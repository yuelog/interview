package model

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"interview/conf"
)

// 这种写法把实例赋值给一个包内变量，这样会不会消耗内存（看类型是指针推测应该不会）
var (
	master *sql.DB
)

func InitSqlModel() {
	initMaster()
}

func initMaster() {
	sqlDriverNameMaster := conf.GetIns().SqlDriverNameMaster
	sqlDataSourceNameMaster := conf.GetIns().SqlDataSourceNameMaster
	database, err := sql.Open(sqlDriverNameMaster, sqlDataSourceNameMaster)
	if err != nil {
		glog.Fatalf("数据库连接错误(master)，driver:%v,source:%v,err:%v \n", sqlDriverNameMaster, sqlDataSourceNameMaster, err)
	}
	master = database
	glog.V(3).Infoln("初始化数据库(master)成功,配置:", sqlDataSourceNameMaster)
}
