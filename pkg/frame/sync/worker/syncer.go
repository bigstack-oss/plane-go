package worker

import (
	"fmt"
	"os"
	"reflect"

	"github.com/bigstack-oss/plane-go/pkg/base/config"
	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/cronjob"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/interact"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/interact/stage"
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/plug"
	"github.com/mohae/deepcopy"
	"go.uber.org/zap"
)

const (
	name       = "name"
	stages     = "stages"
	module     = "interactor"
	interfaces = "interfaces"
)

var (
	configer = config.GetConfiger()
	osExit   = os.Exit
)

type (
	confType = map[string]interface{}
)

type Syncer struct {
	Interact string
	CronJobs []string

	log  *zap.Logger
	logf *zap.SugaredLogger
}

func InitWorker() Worker {
	logger := log.GetLogger(module)

	return &Syncer{
		log:  logger,
		logf: logger.Sugar(),
	}
}

func (r *Syncer) SetStage(interfaceName string, stageIndex int, stageConfig interface{}) {
	stageName := stageConfig.(confType)[name].(string)
	stager := fmt.Sprintf("%s-%s-%d", interfaceName, stageName, stageIndex)
	defer func() {
		panic := recover()
		if panic != nil {
			r.logf.Errorf("stage plugin init failed: %s was not found", stageName)
			osExit(1)
		}
	}()

	plug.Stagers[stager] = deepcopy.Copy(stage.Plugins[stageName]).(plug.Stager)
	plug.Stagers[stager].SetConfig(stageConfig)
	err := plug.Stagers[stager].CheckConfig()
	if err != nil {
		r.logf.Errorf("error details of set stager(%s): %s", stager, err.Error())
		osExit(1)
	}
}

func (r *Syncer) setInteractorConfig(interactor string, pluginConf interface{}) {
	plug.InteractPluggers[interactor].SetConfig(pluginConf)
	err := plug.InteractPluggers[interactor].CheckConfig()
	if err != nil {
		r.logf.Error(err)
		osExit(1)
	}
}

func (r *Syncer) getInterfaceConfs(interactConf interface{}) []interface{} {
	return interactConf.(confType)[interfaces].([]interface{})
}

func (r *Syncer) getInteractorConfig(pluginType string) (string, interface{}) {
	rawConf := configer.Get(pluginType)
	pluginName := rawConf.(confType)[name].(string)
	r.Interact = pluginName
	interactor := fmt.Sprintf("%s-%s", pluginType, pluginName)

	return interactor, rawConf
}

func (r *Syncer) SetInteractor(pluginType string) {
	interactor, conf := r.getInteractorConfig(pluginType)
	plug.InteractPluggers[interactor] = interact.Plugins[r.Interact]

	interfaceConfs := r.getInterfaceConfs(conf)
	for _, interfaceConf := range interfaceConfs {
		interfaceName := interfaceConf.(confType)[name].(string)
		interfaceStages, hasStages := interfaceConf.(confType)[stages].([]interface{})
		if !hasStages {
			continue
		}

		for stageIndex, stageConf := range interfaceStages {
			r.SetStage(interfaceName, stageIndex, stageConf)
		}
	}

	r.setInteractorConfig(interactor, conf)
}

func (r *Syncer) StartInteractor() {
	r.logf.Infof("start interact plugin(%s)", r.Interact)
	go interact.Plugins[r.Interact].DoInteract()
}

func (r *Syncer) StopInteractor() {
	interact.Plugins[r.Interact].Stop()
	r.logf.Infof("stop interact plugin(%s)", r.Interact)
}

func (r *Syncer) getCronnerConfig(pluginType string) map[string]confType {
	cronConfigs := make(map[string]confType)
	rawConfig := reflect.ValueOf(configer.Get(pluginType))
	if !rawConfig.IsValid() {
		return nil
	}

	for i := 0; i < rawConfig.Len(); i++ {
		pluginConfig := rawConfig.Index(i).Interface().(confType)
		pluginName := pluginConfig[name].(string)
		cronName := fmt.Sprintf("%s-%s", pluginType, pluginName)

		cronConfigs[pluginName] = make(confType)
		cronConfigs[pluginName] = pluginConfig
		plug.Cronners[cronName] = cronjob.Plugin[pluginName]
	}

	return cronConfigs
}

func (r *Syncer) setCronnerConfig(pluginType string, cronConfigs map[string]confType) {
	for cronName, cronConfig := range cronConfigs {
		cronner := fmt.Sprintf("%s-%s", pluginType, cronName)
		plug.Cronners[cronner].SetConfig(cronConfig)
		err := plug.Cronners[cronner].CheckConfig()
		if err != nil {
			r.logf.Errorf("failed to set cronjob plugin(%s). error: %s", cronName, err.Error())
			osExit(1)
		}

		r.CronJobs = append(r.CronJobs, cronName)
	}
}

func (r *Syncer) SetCronner(pluginType string) {
	cronConfigs := r.getCronnerConfig(pluginType)
	r.setCronnerConfig(pluginType, cronConfigs)
}

func (r *Syncer) StartCronners() {
	for _, cronName := range r.CronJobs {
		r.logf.Infof("start cronjob plugin(%s)", cronName)
		go cronjob.Plugin[cronName].DoSchedule()
	}
}

func (r *Syncer) StopCronners() {
	for _, cronName := range r.CronJobs {
		cronjob.Plugin[cronName].Stop()
		r.logf.Infof("stop cronjob plugin(%s)", cronName)
	}
}
