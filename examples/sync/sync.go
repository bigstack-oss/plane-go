package main

import (
	"github.com/bigstack-oss/plane-go/pkg/frame/sync/controller"

	_ "github.com/bigstack-oss/plane-go/examples/sync/plugin/cronjob"
	_ "github.com/bigstack-oss/plane-go/examples/sync/plugin/interact/http/pattern/dummy-interact-get"
	_ "github.com/bigstack-oss/plane-go/examples/sync/plugin/interact/http/pattern/dummy-interact-post"
	_ "github.com/bigstack-oss/plane-go/examples/sync/plugin/interact/stage/process"
	_ "github.com/bigstack-oss/plane-go/examples/sync/plugin/interact/stage/request"
	_ "github.com/bigstack-oss/plane-go/examples/sync/plugin/interact/stage/transit"

	_ "github.com/bigstack-oss/plane-go/pkg/frame/sync/plugin/interact/http"
)

func main() {
	c := controller.GetInstance()
	c.InitService()
	c.ActivateService()

	c.Start()

	c.TrapSignals()
	c.TraceStatus()
}
