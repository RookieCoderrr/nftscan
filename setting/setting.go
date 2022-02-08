package setting

import (
	"crypto-colly/common/db"
	"crypto-colly/common/redis"
)

type Config struct {
	Db      db.Config `yaml:"mongo"`
	Redis   redis.Config `yaml:"redis"`
	ChainId uint `yaml:"chain_id"`
}