package redis

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"time"
)

const(
	checkInterval = time.Second * 60
)

type Redis struct {
	pool *redis.Pool
}

type Config struct {
	Host string
	Port string
	Password string
	Db string
}

func InitializeRedisLocalClient(config *Config) *Redis {
	r := Redis{}
	r.pool = &redis.Pool{
		MaxIdle:      20,
		MaxActive:    100,
		IdleTimeout:  180 * time.Second,
		TestOnBorrow: r.check,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Host+":"+config.Port)
			if err != nil {
				panic(err)
			}
			// 使用密码
			if config.Password != "" {
				_, err := c.Do("AUTH", config.Password)
				if err != nil {
					panic(err)
				}
			}
			// 选择db
			if config.Db != "" {
				_, err := c.Do("SELECT", config.Db)
				if err != nil {
					panic(err)
				}
			}
			return c, nil
		},
	}
	return &r
}

func (r *Redis) check(c redis.Conn, t time.Time) error {
	if time.Since(t) > checkInterval {
		_, err := c.Do("PING")
		return err
	}
	return nil
}

func (r *Redis) GetConn() redis.Conn {
	conn := r.pool.Get()
	return conn
}

func (r *Redis) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	conn := r.GetConn()
	defer conn.Close()
	return conn.Do(commandName, args...)
}
func (r *Redis) Test(){
	conn := r.GetConn()
	_, err := conn.Do("Set", "abc", 200)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = redis.Int(conn.Do("Get", "abc"))
	if err != nil {
		fmt.Println("get abc faild :",err)
		return
	}
	fmt.Println("======initialize redis successfully======")
	defer conn.Close()
}
