package config

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	initConfigOnce sync.Once
	_config        *Config
)

type Config struct {
	AppListen  string          `yaml:"app-listen"`
	KeepAlive  int             `yaml:"keep-alive"`
	ServerConf []*ServerConfig `yaml:"server"`
	ClientConf *ClientConfig   `yaml:"client"`
}

func init() {
	initConfigOnce.Do(loadConfig)
}

func loadConfig() {
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "--config config.yaml")
	flag.Parse()

	file, err := os.Open(configFile)
	if err != nil {
		panic(fmt.Sprintf("open config file failed, path:%s, error:%v", configFile, err))
	}
	conf := &Config{}
	decode := yaml.NewDecoder(file)
	if err := decode.Decode(conf); err != nil {
		panic(err)
	}

	conf = _loadConfig(conf)

	for i, srvConf := range conf.ServerConf {
		conf.ServerConf[i] = loadServerConfig(srvConf)
	}

	conf.ClientConf = loadClientConfig(conf.ClientConf)

	_config = conf
}

func _loadConfig(conf *Config) *Config {
	if conf == nil {
		return conf
	}
	if conf.KeepAlive == 0 {
		conf.KeepAlive = 10000
	}
	return conf
}

func GetConfig() *Config {
	if _config == nil {
		panic("config not init")
	}
	return _config
}

func GetServerConfig(name string) *ServerConfig {
	var config = &ServerConfig{}
	if _config == nil {
		panic("config not init")
	}
	for _, conf := range _config.ServerConf {
		if conf.Name == name {
			config = conf
			break
		}
	}
	return config
}

func GetClientConfig() *ClientConfig {
	if _config == nil {
		return &ClientConfig{}
	}
	return _config.ClientConf
}