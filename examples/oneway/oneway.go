package main

import (
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/controller"

	_ "github.com/bigstack-oss/plane-go/examples/oneway/plugin/cronjob"
	_ "github.com/bigstack-oss/plane-go/examples/oneway/plugin/input"
	_ "github.com/bigstack-oss/plane-go/examples/oneway/plugin/output"
	_ "github.com/bigstack-oss/plane-go/examples/oneway/plugin/process"
	_ "github.com/bigstack-oss/plane-go/examples/oneway/plugin/transit"
)

func main() {
	c := controller.GetInstance()
	c.InitService()
	c.ActivateService()

	c.Start()

	c.TrapSignals()
	c.TraceStatus()
}
