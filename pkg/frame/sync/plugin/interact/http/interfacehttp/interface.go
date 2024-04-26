package interfacehttp

import (
	"github.com/gin-gonic/gin"
)

var (
	Plugins = make(map[string]Interface)
)

type Interface interface {
	SetConfig()
	RegisterRouter(*gin.Engine, string, string) error
	AppendStage(string)
}
