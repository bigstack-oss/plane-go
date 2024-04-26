package cronjob

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWrapWithCron(t *testing.T) {
	testVar := 0
	testFunc := func() {
		testVar = 1
	}
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	wrappedFunc := WrapWithCron(ctx, wg, "* * * * * *", testFunc)

	go func() {
		time.Sleep(time.Second * 1)
		cancel()
	}()
	wrappedFunc()
	assert.Equal(t, 1, testVar, "")
}
