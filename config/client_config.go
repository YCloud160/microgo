package config

import "time"

const (
	defaultRequestTimeout          = 5000
	defaultRefreshEndpointInterval = 10000
)

type ClientConfig struct {
	RequestTimeout          int64 `yaml:"request-timeout"`
	RefreshEndpointInterval int64 `yaml:"refresh-endpoint-interval"`
}

func loadClientConfig(conf *ClientConfig) *ClientConfig {
	if conf == nil {
		conf = &ClientConfig{}
	}
	conf.RequestTimeout = getValue(conf.RequestTimeout, 1000, defaultRequestTimeout) * int64(time.Millisecond)
	conf.RefreshEndpointInterval = getValue(conf.RefreshEndpointInterval, 1000, defaultRefreshEndpointInterval)
	return conf
}
