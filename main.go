package main

import (
	"crypto-colly/app"
	cfg "crypto-colly/common/config"
	"crypto-colly/common/db"
	"crypto-colly/common/redis"
	"crypto-colly/setting"
	"flag"
)

func main() {
	var confFile = flag.String("c","config.yml","setting file")
	flag.Parse()
	conf := new(setting.Config)
	cfg.NewConfig(conf).Read(*confFile)
	redisConn := redis.InitializeRedisLocalClient(&conf.Redis)
	dbConn, err := db.NewDb(&conf.Db)
	redisConn.Test()
	if err != nil {
		panic(err)
	}
	app.NewApp(conf,dbConn.GetConn(),redisConn).Do()
}