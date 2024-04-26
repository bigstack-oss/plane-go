package config

import (
	"io"

	"github.com/spf13/viper"
)

var instance Configer

// currently, we only support viper here
// in the future, might can consider to support multiple types of configer
func init() {
	instance = viper.New()
}

func GetConfiger() Configer {
	return instance
}

type Configer interface {
	SetConfigType(string)
	ReadConfig(io.Reader) error
	Get(string) interface{}
	GetInt32(string) int32
}
