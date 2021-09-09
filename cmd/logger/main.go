package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubserve-go/pkg/hubconfig"
	"github.com/wostzone/hubserve-go/pkg/proc"
	"github.com/wostzone/logger/internal"
)

func main() {
	svc := internal.NewLoggerService()
	hubConfig, err := hubconfig.LoadCommandlineConfig("", internal.PluginID, &svc.Config)
	if err != nil {
		logrus.Errorf("ERROR: Start aborted due to error")
		os.Exit(1)
	}
	err = svc.Start(hubConfig)
	if err != nil {
		logrus.Errorf("Logger: Failed to start")
		os.Exit(1)
	}
	proc.WaitForSignal()
	svc.Stop()
	os.Exit(0)
}
