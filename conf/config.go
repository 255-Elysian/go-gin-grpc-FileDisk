package conf

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

var Conf *Config

type Config struct {
	Server   *Server             `yaml:"server"`
	MySQL    *MySQL              `yaml:"mysql"`
	Redis    *Redis              `yaml:"redis"`
	Etcd     *Etcd               `yaml:"etcd"`
	Services map[string]*Service `yaml:"services"`
	Domain   map[string]*Domain  `yaml:"domains"`
	Token    *Token              `yaml:"token"`
	Kafka    *Kafka              `yaml:"kafka"`
	Qiniu    *Qiniu              `yaml:"qiniu"`
}

type Server struct {
	Port    string `yaml:"port"`
	Version string `yaml:"version"`
}

type MySQL struct {
	DriveName string `yaml:"driveName"`
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	Database  string `yaml:"database"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Charset   string `yaml:"charset"`
}

type Redis struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Address  string `yaml:"address"`
}

type Etcd struct {
	Endpoints []string `yaml:"Endpoints"`
}

type Service struct {
	Name        string   `yaml:"name"`
	LoadBalance bool     `yaml:"loadBalance"`
	Addr        []string `yaml:"addr"`
}

type Domain struct {
	Name string `yaml:"name"`
}

type Token struct {
	ShortDuration int `yaml:"shortDuration"`
	LongDuration  int `yaml:"longDuration"`
}

type Kafka struct {
	Topic   []string `yaml:"topic"`
	Broker  []string `yaml:"broker"`
	GroupId []string `yaml:"groupID"`
}

type Qiniu struct {
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
	Bucket    string `yaml:"bucket"`
	Domain    string `yaml:"domain"`
	Zone      string `yaml:"zone"`
}

func InitConfig() {
	workDir, _ := os.Getwd()
	viper.AddConfigPath(workDir + "/conf")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	err := viper.Unmarshal(&Conf)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
