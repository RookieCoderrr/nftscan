package db

import (
	"crypto-colly/models"
	"github.com/ethereum/go-ethereum/log"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Db struct {
	conn *gorm.DB
}

type Config struct {
	Host string
	Port string
	User string
	Password string
	DbName string
}

func NewDb(options *Config) (*Db, error) {
	var (
		db   *Db
		conn *gorm.DB
		err  error
	)
	db = &Db{}

	conn, err = gorm.Open("mysql",
		options.User+":"+options.Password+"@tcp("+options.Host+":"+options.Port+")/"+options.DbName+
			"?charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true")
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	db.conn = conn
	conn.AutoMigrate(&models.NFTTransaction{}, &models.NFTAsset{}, &models.NFTTransfer{},&models.NFTContract{})

	return db, nil
}

func (d *Db) GetConn() *gorm.DB {
	return d.conn
}
