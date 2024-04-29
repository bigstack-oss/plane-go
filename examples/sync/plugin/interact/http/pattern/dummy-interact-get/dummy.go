package dummy

import (
	"errors"
	"net/http"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/interact/http/interfacehttp"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/plug"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	module   = "dummy-interact-get"
	response = "dummy get response"
)

var (
	get = "GET"
)

type DummyHandler struct {
	stages []plug.Stager

	log  *zap.Logger
	logf *zap.SugaredLogger
}

func init() {
	interfacehttp.Plugins[module] = &DummyHandler{}
}

func (m *DummyHandler) SetConfig() {
	m.log = log.GetLogger(module)
	m.logf = m.log.Sugar()
}

func (m *DummyHandler) RegisterRouter(router *gin.Engine, method string, path string) error {
	switch method {
	case get:
		router.GET(path, m.get)
	default:
		m.logf.Errorf("unsupported method(%s) detected in pattern(%s)", method, path)
		return errors.New("register unsupported REST method")
	}

	return nil
}

func (m *DummyHandler) AppendStage(stage string) {
	m.stages = append(m.stages, plug.Stagers[stage])
}

func (m *DummyHandler) get(g *gin.Context) {
	for _, stage := range m.stages {
		stage.Execute(&protocol.Job{})
	}

	g.JSON(http.StatusOK, response)
}
