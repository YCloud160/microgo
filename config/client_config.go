package config

import "time"

const (
	defaultRequestTimeout = 5000
)

type ClientConfig struct {
	RequestTimeout int64 `yaml:"request-timeout"`
}

func loadClientConfig(conf *ClientConfig) *ClientConfig {
	if conf == nil {
		conf = &ClientConfig{}
	}
	conf.RequestTimeout = getValue(conf.RequestTimeout, 1000, defaultRequestTimeout) * int64(time.Millisecond)
	return conf
}
