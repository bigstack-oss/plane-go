package stage

import "github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/plug"

var (
	Plugins = make(map[string]plug.Stager)
)
