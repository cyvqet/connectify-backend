package ioc

import (
	"github.com/cyvqet/connectify/internal/repository/dao"
	"github.com/spf13/viper"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	type DBConfig struct {
		DSN string `yaml:"dsn"`
	}
	var dbConfig DBConfig
	err := viper.UnmarshalKey("db.mysql", &dbConfig)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.Open(dbConfig.DSN))
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
