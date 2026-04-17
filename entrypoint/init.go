package entrypoint

import (
  "fmt"
	. "users/version"
	. "users/global"

	"github.com/spf13/viper"
)

func Init() {
  Update()
	updateversion()
}

func updateversion() {
setingskey := fmt.Sprintf("%s.entrypointversion", MicroServiceName)
viper.Set(setingskey, VERSIONPLUGIN)
}
