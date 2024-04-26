package output

import (
	"context"
	"sync"

	"github.com/bigstack-oss/plane-go/pkg/base/protocol"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin"
	"github.com/bigstack-oss/plane-go/pkg/frame/oneway/plugin/plug"
)

var (
	Plugin = make(map[string]Output)
)

type Output interface {
	plug.Parter
	DoOutput()
}

func WrapWithSingleMsgLoop(ctx context.Context, wg *sync.WaitGroup, group string, coreFunc func(protocol.Job) error) func() {
	return func() {
		wg.Add(1)
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case message, isChnOpen := <-plugin.P2OChan[group]:
				switch isChnOpen {
				case true:
					err := coreFunc(message)
					if err != nil {
						continue
					}

				case false:
					return
				}
			}
		}
	}
}
