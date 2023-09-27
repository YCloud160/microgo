package microgo

import (
	"github.com/YCloud160/microgo/config"
	discovery2 "github.com/YCloud160/microgo/internal/discovery"
)

type Discovery interface {
	QueryRoute(name string) ([]string, error)
}

var discovery Discovery

func initDiscovery(conf *config.Registry) {
	switch conf.Name {
	case "micro-route":
		discovery = discovery2.NewMicroDiscovery(conf.Data["host"])
	}
}
