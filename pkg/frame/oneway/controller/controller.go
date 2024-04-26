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
	"time"

	"github.com/bigstack-oss/plane-go/pkg/base/config"
	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/monitoring"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/worker"

	// auto setup the maxprocs with the request/limit cpu
	// please don't remove it unless you know the cpu factors between Golang and cgroup
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

const (
	module   = "controller"
	runMode  = "oneTimeExec"
	chanSize = "channelSize"
	yamlConf = "yaml"
)

var (
	instance *controller

	wg   sync.WaitGroup = sync.WaitGroup{}
	once sync.Once

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

	isWorkerCompleted bool
	isOneTimeExec     bool
	signalChan        chan os.Signal

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

	configer.SetConfigType(yamlConf)
	err = configer.ReadConfig(bytes.NewBuffer(content))
	if err != nil {
		return err
	}

	return nil
}

func (c *controller) initControllerParams() {
	c.wg = &wg
	c.ctx, c.cancel = context.WithCancel(context.Background())

	c.signalChan = make(chan os.Signal, 1)
	c.isOneTimeExec = configer.Get(runMode).(bool)
	c.isWorkerCompleted = false
}

func (c *controller) initPluginParams() {
	plugin.Service = strings.TrimSuffix(conf, filepath.Ext(conf))

	plugin.IsOneTimeExec = c.isOneTimeExec
	plugin.ChanSize = configer.GetInt32(chanSize)

	plugin.Metrics = &plugin.Metric{}
	plugin.Records = &plugin.Record{}
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
	c.Worker.SetParter(plugin.Input)
	c.Worker.SetParter(plugin.Transit)
	c.Worker.SetParter(plugin.Process)
	c.Worker.SetParter(plugin.Output)
	c.Worker.SetCronner(plugin.CronJob)
}

func (c *controller) Start() {
	c.log.Info("activating worker")
	c.Worker.StartParters()
	c.Worker.StartCronners()
	c.wg.Add(1)
}

func (c *controller) Stop() {
	c.Worker.StopParters()
	c.Worker.StopCronners()
	c.wg.Done()
	c.log.Info("worker has been stopped")
}

func (c *controller) Restart() {
	c.log.Info("restarting worker")
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
		signal.Notify(c.signalChan, syscall.SIGHUP, syscall.SIGTERM)
		c.log.Info("signal registered: SIGTERM, SIGUSR1")

		for sig := range c.signalChan {
			switch sig {
			case syscall.SIGHUP:
				c.log.Info("Signal SIGHUP received, restarting service...")
				c.Restart()
			case syscall.SIGTERM:
				c.log.Info("Signal SIGTERM received, stopping service ...")
				c.Stop()
			}
		}
	}()
}

func (c *controller) MonitorService() {
	c.Monitor.SetReportTunnel(c.isOneTimeExec)
	c.Monitor.TraceMetric()
}

func (c *controller) TraceStatus() {
	go func() {
		if !c.isOneTimeExec {
			return
		}

		for {
			c.isWorkerCompleted = c.Worker.GetStatus()
			if c.isWorkerCompleted {
				c.Stop()
				break
			}

			time.Sleep(3 * time.Second)
		}
	}()

	c.MonitorService()

	c.wg.Wait()
	c.log.Info("worker is done, ready to exit process")
}
