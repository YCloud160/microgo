package config

import "time"

const (
	maxInvokeNum = 10000
)

type ServerConfig struct {
	Name          string `yaml:"name"`
	IP            string `yaml:"ip"`
	Port          string `yaml:"port"`
	InvokeTimeout int64  `yaml:"invoke-timeout"`
	MaxInvoke     int64  `yaml:"max-invoke"`
}

func loadServerConfig(conf *ServerConfig) *ServerConfig {
	if conf == nil {
		return conf
	}
	conf.InvokeTimeout = getValue(conf.InvokeTimeout, 1000, 0) * int64(time.Millisecond)
	conf.MaxInvoke = getValue(conf.MaxInvoke, 1, maxInvokeNum)
	return conf
}

func getValue(inputVal, compareVal, defaultVal int64) int64 {
	if inputVal < compareVal {
		return defaultVal
	}
	return inputVal
}
