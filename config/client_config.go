package config

import "time"

const (
	defaultRequestTimeout = int(time.Millisecond * 5000)
)

type ClientConfig struct {
	RequestTimeout int `yaml:"request-timeout"`
}

func loadClientConfig(conf *ClientConfig) *ClientConfig {
	if conf == nil {
		conf = &ClientConfig{}
	}
	conf.RequestTimeout = conf.RequestTimeout * int(time.Millisecond)
	if conf.RequestTimeout < 1000 {
		conf.RequestTimeout = defaultRequestTimeout
	}
	return conf
}
