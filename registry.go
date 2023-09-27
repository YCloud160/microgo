package microgo

import (
	"github.com/YCloud160/microgo/config"
	registry2 "github.com/YCloud160/microgo/internal/registry"
)

type Registry interface {
	Register(name string, addr string) error
	UnRegister(name string, addr string) error
	KeepAlive(name string, addr string) error
}

var registry Registry

func initRegistry(conf *config.Registry) {
	switch conf.Name {
	case "micro-route":
		registry = registry2.NewMicroRegistry(conf.Data["host"])
	}
}
