package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/pkg/hubconfig"
	"github.com/wostzone/hubapi/pkg/plugin"
	"github.com/wostzone/logger/internal"
)

var pluginConfig = &internal.WostLoggerConfig{}

func main() {
	hubConfig, err := hubconfig.SetupConfig("", internal.PluginID, pluginConfig)

	svc := internal.WostLogger{}
	err = svc.Start(hubConfig, pluginConfig)
	if err != nil {
		logrus.Errorf("Logger: Failed to start")
		os.Exit(1)
	}
	plugin.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
