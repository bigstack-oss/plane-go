package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfiger(t *testing.T) {
	testConfiger := GetConfiger()
	assert.NotEqual(t, nil, testConfiger, "failed to get configer")
}
