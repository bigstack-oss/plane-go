package plug

import "github.com/bigstack-oss/plane-go/pkg/base/protocol"

var Stagers = make(map[string]Stager)

type ConfigSetter interface {
	SetConfig(interface{})
}

type Stager interface {
	ConfigSetter
	ConfigChecker
	Execute(*protocol.Job) (bool, error)
}
