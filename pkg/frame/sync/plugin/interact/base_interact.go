package interact

import (
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/plug"
)

var (
	Plugins = make(map[string]Interactor)
)

type Interactor interface {
	plug.InteractPlugger
	DoInteract()
}
