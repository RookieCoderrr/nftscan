package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type (
	Config struct {
		Data interface{}
	}
)

func NewConfig(data interface{}) *Config{
	return &Config{
		Data: data,
	}
}
func (c *Config) Read(file string) error{
	var err error
	if err = c.ReadYaml(file); err != nil {
		fmt.Println("real .yml setting error",err)
		return err
	}
	return nil
}

func (c *Config) ReadYaml(file string) error {
	var (
		err  error
		data []byte
	)

	data, err = ioutil.ReadFile(file)
	if err != nil {
		fmt.Sprintf("Read yml setting: %s",err)
		return err
	}

	err = yaml.Unmarshal([]byte(data), c.Data)
	if err != nil {
		fmt.Sprintf("Read yml setting: %s",err)
		return err
	}

	return nil
}



