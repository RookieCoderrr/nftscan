package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Db struct {
	conn *mongo.Client
}

type Config struct {
	Host string
	Port string
	Database string
}

func InitializeMongoLocalClient(ctx context.Context, config *Config) (*Db, error) {
	var (
		db *Db
		conn *mongo.Client
		err error
	)
	db = &Db{}
	var clientOptions *options.ClientOptions
	clientOptions = options.Client().ApplyURI("mongodb://" + config.Host + ":" + config.Port + "/" + config.Database)
	conn, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("connect mongo error")
		return nil, err
	}
	err = conn.Ping(ctx, nil)
	if err != nil {
		fmt.Println("ping mongo error")
		return nil, err
	}
	db.conn = conn
	return db,nil
}

func (d *Db) GetConn() *mongo.Client{
	return d.conn
}