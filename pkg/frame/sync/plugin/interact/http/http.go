package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/interact"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/interact/http/interfacehttp"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/plug"
	"github.com/gin-gonic/gin"
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const (
	module = "http"
)

var (
	Router *gin.Engine
)

type Http struct {
	ctx    context.Context
	cancel context.CancelFunc

	listener plug.Listener
	router   *gin.Engine
	config

	log  *zap.Logger
	logf *zap.SugaredLogger
}

type config struct {
	Name       string `validate:"required"`
	Address    string `validate:"required"`
	Port       int    `validate:"required"`
	Interfaces []Interface
}

type Interface struct {
	Name    string `validate:"required"`
	Method  string `validate:"required"`
	Path    string `validate:"required"`
	Timeout int
	Stages  []Stage
}

type Stage struct {
	Name string
}

func init() {
	registerModule()
}

func registerModule() {
	interact.Plugins[module] = &Http{}
}

func (h *Http) setServer() {
	socket := fmt.Sprintf("%s:%d", h.Address, h.Port)
	h.listener = &http.Server{
		Addr:    socket,
		Handler: h.router,
	}
}

func (h *Http) setRouter() {
	gin.DefaultWriter = ioutil.Discard
	h.router = gin.New()
	h.router.Use(gin.Recovery())
}

func (h *Http) setStageOrder(interfaceName string, stages []Stage) {
	for i, s := range stages {
		stage := fmt.Sprintf("%s-%s-%d", interfaceName, s.Name, i)
		interfacehttp.Plugins[interfaceName].AppendStage(stage)
	}
}

func (h *Http) setStages() {
	for _, i := range h.Interfaces {
		err := interfacehttp.Plugins[i.Name].RegisterRouter(h.router, i.Method, i.Path)
		if err != nil {
			h.logf.Errorf("fail to register router. error details: %s", err.Error())
		}

		h.setStageOrder(i.Name, i.Stages)
	}
}

func (h *Http) SetConfig(conf interface{}) {
	_ = mapstructure.Decode(conf, &h.config)
	h.ctx, h.cancel = context.WithCancel(context.Background())

	h.setRouter()
	h.setStages()
	h.setServer()

	h.log = log.GetLogger(module)
	h.logf = h.log.Sugar()
}

func (h *Http) CheckConfig() error {
	return validator.New().Struct(h.config)
}

func (h *Http) DoInteract() {
	err := h.listener.ListenAndServe()
	if err != nil {
		h.logf.Errorf("error details of start http listener: %s", err.Error())
	}
}

func (h *Http) Stop() {
	if err := h.listener.Shutdown(h.ctx); err != nil {
		h.logf.Errorf("failed to stop interact plugin(%s). error: %s", module, err.Error())
	}
}
