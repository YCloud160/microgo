package config

const (
	maxInvokeNum = 10000
)

type ServerConfig struct {
	Name          string `yaml:"name"`
	IP            string `yaml:"ip"`
	Port          string `yaml:"port"`
	InvokeTimeout int    `yaml:"invoke-timeout"`
	MaxInvoke     int    `yaml:"max-invoke"`
}

func loadServerConfig(conf *ServerConfig) *ServerConfig {
	if conf == nil {
		return conf
	}
	if conf.MaxInvoke <= 0 {
		conf.MaxInvoke = maxInvokeNum
	}
	return conf
}
