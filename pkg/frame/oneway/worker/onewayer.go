package worker

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/bigstack-oss/plane-go/pkg/base/config"
	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/cronjob"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/input"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/output"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/plug"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/process"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/transit"
	"github.com/mohae/deepcopy"
	"go.uber.org/zap"
)

const (
	name   = "name"
	module = "onewayer"
)

var (
	configer = config.GetConfiger()
	osExit   = os.Exit
)

type Onewayer struct {
	Inputs    []string
	Transits  []string
	Processes []string
	Outputs   []string
	Crons     []string

	log  *zap.Logger
	logf *zap.SugaredLogger
}

func InitWorker() Worker {
	logger := log.GetLogger(module)

	return &Onewayer{
		log:  logger,
		logf: logger.Sugar(),
	}
}

func (o *Onewayer) getParterConfigs(pluginType string) ([]string, []interface{}) {
	parterNames := []string{}
	rawConfigs := configer.Get(pluginType).([]interface{})

	for i, rawConfig := range rawConfigs {
		subConfig := (rawConfig.(map[string]interface{}))
		pluginName := subConfig[name].(string)
		pluginGroup := subConfig["group"].(string)
		parterName := strings.Join([]string{pluginType, pluginName, strconv.Itoa(i)}, "-")

		switch pluginType {
		case plugin.Input:
			plug.Parters[parterName] = deepcopy.Copy(input.Plugin[pluginName]).(input.Input)
			input.Plugin[parterName] = plug.Parters[parterName].(input.Input)

			o.Inputs = append(o.Inputs, parterName)
			plugin.I2TChan[pluginGroup] = make(chan []byte, plugin.ChanSize)

		case plugin.Transit:
			plug.Parters[parterName] = deepcopy.Copy(transit.Plugin[pluginName]).(transit.Transit)
			transit.Plugin[parterName] = plug.Parters[parterName].(transit.Transit)

			o.Transits = append(o.Transits, parterName)
			plugin.T2PChan[pluginGroup] = make(chan protocol.Job, plugin.ChanSize)

		case plugin.Process:
			plug.Parters[parterName] = deepcopy.Copy(process.Plugin[pluginName]).(process.Process)
			process.Plugin[parterName] = plug.Parters[parterName].(process.Process)

			o.Processes = append(o.Processes, parterName)
			plugin.P2OChan[pluginGroup] = make(chan protocol.Job, plugin.ChanSize)

		case plugin.Output:
			plug.Parters[parterName] = deepcopy.Copy(output.Plugin[pluginName]).(output.Output)
			output.Plugin[parterName] = plug.Parters[parterName].(output.Output)

			o.Outputs = append(o.Outputs, parterName)
		}

		parterNames = append(parterNames, parterName)
	}

	return parterNames, rawConfigs
}

func (o *Onewayer) setParterConfig(parterNames []string, parterConfigs []interface{}) {
	for i, parterName := range parterNames {
		plug.Parters[parterName].SetConfig(parterConfigs[i])
		err := plug.Parters[parterName].CheckConfig()
		if err != nil {
			o.logf.Errorf("failed to set plugin: %s", parterName)
			o.logf.Errorf("error details: %s", err.Error())
			osExit(1)
		}
	}
}

func (o *Onewayer) SetParter(pluginType string) {
	parterNames, parterConfigs := o.getParterConfigs(pluginType)
	o.setParterConfig(parterNames, parterConfigs)
}

func (o *Onewayer) StartParters() {
	for _, name := range o.Inputs {
		o.logf.Infof("start input plugin(%s)", name)
		go input.Plugin[name].DoInput()
	}

	for _, name := range o.Transits {
		o.logf.Infof("start transit plugin(%s)", name)
		go transit.Plugin[name].DoTransit()
	}

	for _, name := range o.Processes {
		o.logf.Infof("start process plugin(%s)", name)
		go process.Plugin[name].DoProcess()
	}

	for _, name := range o.Outputs {
		o.logf.Infof("start output plugin(%s)", name)
		go output.Plugin[name].DoOutput()
	}
}

func (o *Onewayer) StopParters() {
	for _, name := range o.Inputs {
		input.Plugin[name].Stop()
		o.logf.Infof("stop input plugin(%s)", name)
	}

	for _, name := range o.Transits {
		transit.Plugin[name].Stop()
		o.logf.Infof("stop transit plugin(%s)", name)
	}

	for _, name := range o.Processes {
		process.Plugin[name].Stop()
		o.logf.Infof("stop process plugin(%s)", name)
	}

	for _, name := range o.Outputs {
		output.Plugin[name].Stop()
		o.logf.Infof("stop output plugin(%s)", name)
	}
}

func (o *Onewayer) getCronnerConfig(pluginType string) map[string]map[string]interface{} {
	cronConfigs := make(map[string]map[string]interface{})
	rawConfig := reflect.ValueOf(configer.Get(pluginType))
	if !rawConfig.IsValid() {
		return map[string]map[string]interface{}{}
	}

	for i := 0; i < rawConfig.Len(); i++ {
		pluginConfig := rawConfig.Index(i).Interface().(map[string]interface{})
		pluginName := pluginConfig[name].(string)
		cronName := strings.Join([]string{pluginType, pluginName}, "-")

		cronConfigs[pluginName] = make(map[string]interface{})
		cronConfigs[pluginName] = pluginConfig
		plug.Cronners[cronName] = cronjob.Plugin[pluginName]
	}

	return cronConfigs
}

func (o *Onewayer) setCronnerConfig(pluginType string, cronConfigs map[string]map[string]interface{}) {
	var cronName string
	defer func() {
		if err := recover(); err != nil {
			o.logf.Errorf("cronjob plugin(%s) was not defined", cronName)
			osExit(1)
		}
	}()

	for cronName, cronConfig := range cronConfigs {
		cronner := strings.Join([]string{pluginType, cronName}, "-")
		plug.Cronners[cronner].SetConfig(cronConfig)
		err := plug.Cronners[cronner].CheckConfig()
		if err != nil {
			o.logf.Errorf("failed to set cronjob plugin(%s). error: %s", cronName, err.Error())
			osExit(1)
		}

		o.Crons = append(o.Crons, cronName)
	}
}

func (o *Onewayer) SetCronner(pluginType string) {
	cronConfigs := o.getCronnerConfig(pluginType)
	o.setCronnerConfig(pluginType, cronConfigs)
}

func (o *Onewayer) StartCronners() {
	for _, cron := range o.Crons {
		o.logf.Infof("start cronjob plugin(%s)", cron)
		go cronjob.Plugin[cron].DoSchedule()
	}
}

func (o *Onewayer) StopCronners() {
	for _, cron := range o.Crons {
		cronjob.Plugin[cron].Stop()
		o.logf.Infof("stop cronjob plugin(%s)", cron)
	}
}

func (o *Onewayer) GetStatus() bool {
	isInputDone := *plugin.InputDone == int64(len(o.Inputs))
	isTransitDone := *plugin.TransitDone == int64(len(o.Transits))
	isProcessDone := *plugin.ProcessDone == int64(len(o.Processes))
	isOutputDone := *plugin.OutputDone == int64(len(o.Outputs))

	return isInputDone && isTransitDone && isProcessDone && isOutputDone
}
