package controller

import (
	"context"
	"errors"
	"io"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/bigstack-oss/plane-go/pkg/base/config"
	"github.com/bigstack-oss/plane-go/pkg/base/log"
	"github.com/bigstack-oss/plane-go/pkg/base/monitoring"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

const (
	input   = "testInput"
	transit = "testTransit"
	process = "testProcess"
	output  = "testOutput"
	cronJob = "cronJob"

	successConf            = "../../../example/oneway/oneway-conf.yaml"
	failureNotFoundConf    = "can_not_find_this_file.yaml"
	failureReadConfContent = "../../../example/oneway/oneway-conf.yaml"

	start = "start"
	stop  = "stop"
)

var (
	isTestWorkerCompleted bool
	_                     = func() bool {
		testing.Init()
		return true
	}()
)

type testWorker struct {
	Input    string
	Transit  string
	Process  string
	Output   string
	CronJobs []string

	testWokrerStatus string
	testCronJbStatus string
}

var tester *testWorker

func (t *testWorker) SetParter(pluginType string) {
	switch pluginType {
	case plugin.Input:
		t.Input = input
	case plugin.Transit:
		t.Transit = transit
	case plugin.Process:
		t.Process = process
	case plugin.Output:
		t.Output = output
	}
}

func (t *testWorker) StartParters() { t.testWokrerStatus = start }

func (t *testWorker) StopParters() { t.testWokrerStatus = stop }

func (t *testWorker) SetCronner(pluginType string) {
	t.CronJobs = append(t.CronJobs, cronJob)
}

func (t *testWorker) StartCronners() { t.testCronJbStatus = start }

func (t *testWorker) StopCronners() { t.testCronJbStatus = stop }

func (t *testWorker) GetStatus() bool {
	return isTestWorkerCompleted
}

type testConfiger struct{}

func (tcfgr *testConfiger) SetConfigType(string) {}

func (tcfgr *testConfiger) ReadConfig(ir io.Reader) error { return errors.New("use for error cases") }

func (tcfgr *testConfiger) Get(confKey string) interface{} { return confKey }

func (tcfgr *testConfiger) GetInt32(confKey string) int32 { return 0 }

func TestController(t *testing.T) {
	GetInstance()

	assert.NotEqual(t, nil, instance, "failed to get controller instance")
	assert.NotEqual(t, nil, instance.Worker, "failed to get worker instance")
}

func TestInitService(t *testing.T) {
	var testReturnCode int
	osExit = func(code int) {
		testReturnCode = code
	}

	defer func() {
		osExit = os.Exit
		conf = successConf
	}()

	conf = successConf
	instance.InitService()
	assert.NotEmpty(t, instance.log, "failed to init logger")
	assert.NotEqual(t, nil, instance.ctx, "failed to init ctx")
	assert.NotEqual(t, nil, instance.wg, "failed to init wg")
	assert.NotEqual(t, nil, plugin.Metrics, "failed to init plugin metrics")
	assert.NotEqual(t, nil, plugin.Records, "failed to init plugin records")

	conf = ""
	instance.InitService()
	assert.Equal(t, 1, testReturnCode, "failed to get error return code")
}

func TestLoadConfFile(t *testing.T) {
	var err error
	configer = &testConfiger{}
	defer func() { configer = config.GetConfiger() }()

	err = instance.loadConfig("")
	assert.Equal(t, "conf file is required, please specify the path of conf file", err.Error(), "failed to get conf error")

	err = instance.loadConfig(failureNotFoundConf)
	assert.NotEqual(t, nil, err, "failed to get conf error")

	err = instance.loadConfig(failureReadConfContent)
	assert.Equal(t, "use for error cases", err.Error(), "failed to get conf error")
}

func TestActivateService(t *testing.T) {
	tester = &testWorker{}
	instance.Worker = tester

	instance.ActivateService()
	assert.Equal(t, input, tester.Input, "failed to set input plugin")
	assert.Equal(t, transit, tester.Transit, "failed to set transit plugin")
	assert.Equal(t, process, tester.Process, "failed to set process plugin")
	assert.Equal(t, output, tester.Output, "failed to set output plugin")
	assert.Equal(t, cronJob, tester.CronJobs[0], "failed to set cronjob plugin")
}

func TestStartService(t *testing.T) {
	tester = &testWorker{}
	instance.Worker = tester

	instance.Start()
	assert.Equal(t, start, tester.testWokrerStatus, "failed to start worker")
	assert.Equal(t, start, tester.testCronJbStatus, "failed to start cronjob")
	instance.wg.Done()
}

func TestRestartService(t *testing.T) {
	tester = &testWorker{}
	instance.Worker = tester
	instance.wg.Add(1)

	instance.log = log.GetLogger(module)
	instance.ctx, instance.cancel = context.WithCancel(context.Background())
	instance.wg = &wg
	instance.signalChan = make(chan os.Signal, 1)
	tester.testWokrerStatus = stop
	tester.testCronJbStatus = stop

	instance.Restart()
	assert.Equal(t, start, tester.testWokrerStatus, "failed to restart worker")
	assert.Equal(t, start, tester.testCronJbStatus, "failed to restart cronjob")
}

func TestStopService(t *testing.T) {
	tester = &testWorker{}
	instance.Worker = tester
	tester.testWokrerStatus = start
	tester.testCronJbStatus = start
	instance.wg.Add(1)

	instance.Stop()
	assert.Equal(t, stop, tester.testWokrerStatus, "failed to stop worker")
	assert.Equal(t, stop, tester.testCronJbStatus, "failed to stop cronjob")
}

func TestTrapSignalsSIGTERM(t *testing.T) {
	tester = &testWorker{}
	instance.Worker = tester
	tester.testWokrerStatus = start
	tester.testCronJbStatus = start

	instance.TrapSignals()
	instance.signalChan <- syscall.SIGTERM

	time.Sleep(2 * time.Second)
	assert.Equal(t, stop, tester.testWokrerStatus, "failed to stop worker by SIGTERM")
	assert.Equal(t, stop, tester.testCronJbStatus, "failed to stop cronjob by SIGTERM")
}

func TestTrapSignalsSIGHUP(t *testing.T) {
	tester = &testWorker{}
	instance.Worker = tester
	tester.testWokrerStatus = stop
	tester.testCronJbStatus = stop

	instance.TrapSignals()
	instance.signalChan <- syscall.SIGHUP

	time.Sleep(2 * time.Second)
	assert.Equal(t, start, tester.testWokrerStatus, "failed to stop worker by SIGHUP")
	assert.Equal(t, start, tester.testCronJbStatus, "failed to stop cronjob by SIGHUP")
}

func TestTraceStatusIsOneTimeExec(t *testing.T) {
	tester = &testWorker{}
	instance.Worker = tester
	instance.isOneTimeExec = true
	isTestWorkerCompleted = false
	instance.wg.Add(1)

	go func() {
		time.Sleep(2 * time.Second)
		assert.Equal(t, false, instance.isWorkerCompleted, "failed to sync worker status to false")
	}()

	go func() {
		time.Sleep(3 * time.Second)
		isTestWorkerCompleted = true
	}()

	instance.TraceStatus()
	assert.Equal(t, true, instance.isWorkerCompleted, "failed to sync worker status to true")
}

func TestTraceStatusIsNotOneTimeExec(t *testing.T) {
	tester = &testWorker{}
	instance.Worker = tester
	instance.isOneTimeExec = false
	instance.isWorkerCompleted = false
	instance.wg.Add(1)

	go func() {
		time.Sleep(2 * time.Second)
		instance.wg.Done()
	}()

	monitoring.MetricRegistry = prometheus.NewRegistry()
	instance.TraceStatus()
	assert.Equal(t, false, instance.isWorkerCompleted, "failed to sync worker status to false")
}
