package setup

import (
	"fmt"
	"remote_server/config"
	"remote_server/model"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite"
)

func GormSqlite() *gorm.DB {
	fmt.Println("init sqlite3")
	var err error
	DB, err := gorm.Open("sqlite", config.Config.DBPath)
	if err != nil {
		panic(err)
	}
	//允许单表创建
	DB.SingularTable(true)
	//关闭sql语句日志
	DB.LogMode(true)
	DB.DB().SetMaxIdleConns(10)
	DB.DB().SetMaxOpenConns(100)
	// SetConnMaxLifetiment 设置连接的最大可复用时间。
	DB.DB().SetConnMaxLifetime(time.Hour)
	DB.AutoMigrate(&model.Device{}, &model.Connectioned{})

	return DB
}
