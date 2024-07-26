package microgo

import (
	"fmt"
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

func GetNodes(srvName string) ([]string, error) {
	if discovery == nil {
		return []string{}, fmt.Errorf("discovery not init")
	}
	return discovery.QueryRoute(srvName)
}
