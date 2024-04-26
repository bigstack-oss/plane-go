package controller

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/bigstack-oss/plane-go/pkg/base/config"
	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/monitoring"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/worker"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

const (
	module        = "controller"
	isOneTimeExec = false
)

var (
	instance *controller

	once sync.Once
	wg   sync.WaitGroup = sync.WaitGroup{}

	configer = config.GetConfiger()
	conf     string

	logLevel int

	osExit = os.Exit
)

type controller struct {
	worker.Worker

	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	signalChannel chan os.Signal

	log  *zap.Logger
	logf *zap.SugaredLogger

	monitoring.Monitor
}

func init() {
	flag.StringVar(&conf, "conf", "", "")
	flag.IntVar(&logLevel, "log-level", 2, "")
	flag.Parse()

	log.SetLogLevel(logLevel)
}

func GetInstance() *controller {
	once.Do(func() {
		instance = &controller{
			Worker: worker.InitWorker(),
		}
	})

	return instance
}

func (c *controller) loadConfig(conf string) error {
	if conf == "" {
		return errors.New("conf file is required, please specify the path of conf file")
	}

	content, err := ioutil.ReadFile(conf)
	if err != nil {
		return err
	}

	configer.SetConfigType("yaml")
	err = configer.ReadConfig(bytes.NewBuffer(content))
	if err != nil {
		return err
	}

	return nil
}

func (c *controller) initControllerParams() {
	c.wg = &wg
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.signalChannel = make(chan os.Signal, 1)
}

func (c *controller) initPluginParams() {
	plugin.Service = strings.TrimSuffix(conf, filepath.Ext(conf))
	plugin.Metrics = &plugin.Metric{}
}

func (c *controller) InitService() {
	c.log = log.GetLogger(module)
	c.logf = c.log.Sugar()

	err := c.loadConfig(conf)
	if err != nil {
		c.logf.Errorf("failed to load conf from '%s'. error details: %s", conf, err.Error())
		osExit(1)
	}

	c.initControllerParams()
	c.initPluginParams()
}

func (c *controller) ActivateService() {
	c.Worker.SetInteractor(plugin.Interact)
	c.Worker.SetCronner(plugin.CronJob)
}

func (c *controller) Start() {
	c.log.Info("activating all workers")
	c.Worker.StartInteractor()
	c.Worker.StartCronners()
	c.wg.Add(1)
}

func (c *controller) Stop() {
	c.Worker.StopInteractor()
	c.Worker.StopCronners()
	c.wg.Done()
	c.log.Info("all tasks have been stopped. stopping service.")
}

func (c *controller) Restart() {
	c.log.Info("restarting all workers")
	defer c.wg.Done()
	c.wg.Add(1)
	c.Stop()

	c = GetInstance()
	c.InitService()
	c.ActivateService()
	c.Start()
	c.TrapSignals()
}

func (c *controller) TrapSignals() {
	go func() {
		signal.Notify(c.signalChannel, syscall.SIGHUP, syscall.SIGTERM)
		c.log.Info("signal registered: SIGTERM, SIGUSR1")

		for sig := range c.signalChannel {
			switch sig {
			case syscall.SIGHUP:
				c.log.Info("SIGHUP received, restarting service...")
				c.Restart()
			case syscall.SIGTERM:
				c.log.Info("SIGTERM received, stopping service ...")
				c.Stop()
			}
		}
	}()
}

func (c *controller) MonitorService() {
	c.Monitor.SetReportTunnel(isOneTimeExec)
	c.Monitor.TraceMetric()
}

func (c *controller) TraceStatus() {
	c.MonitorService()

	c.wg.Wait()
	c.log.Info("controller and workers are done")
}
