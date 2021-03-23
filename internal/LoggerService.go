package internal

import (
	"os"
	"path"

	"github.com/sirupsen/logrus"
	hubapi "github.com/wostzone/hubapi/api"
	"github.com/wostzone/hubapi/pkg/hubclient"
	"github.com/wostzone/hubapi/pkg/hubconfig"
)

// PluginID is the default ID of the WoST Logger plugin
const PluginID = "logger"

// WostLoggerConfig with logger plugin configuration
// map of topic -> file
type WostLoggerConfig struct {
	LogsFolder string   `yaml:"logsFolder"` // folder to use for logging
	ThingIDs   []string `yaml:"thingIDs"`   // thing IDs to log
}

// WostLogger is a hub plugin for recording messages to the hub
// By default it logs messages by ThingID, eg each Thing has a log file
type WostLogger struct {
	loggerConfig  WostLoggerConfig
	hubConfig     *hubconfig.HubConfig
	hubConnection hubapi.IHubClient
	loggers       map[string]*logrus.Logger // map of thing ID to logger
}

// handleMessage receives and records a topic message
func (wlog *WostLogger) logToFile(thingID string, msgType string, payload []byte, sender string) {
	logrus.Infof("Received message of type '%s' about Thing %s", msgType, thingID)
	// var err error

	if wlog.loggers == nil {
		logrus.Errorf("logToFile called after logger has stopped")
		return
	}

	// use logrus for logging
	logger := wlog.loggers[thingID]
	if logger == nil {
		logsFolder := wlog.loggerConfig.LogsFolder
		filePath := path.Join(logsFolder, thingID+".log")

		logger = logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000-0700",
			PrettyPrint:     true,
		})
		fileHandle, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)
		if err != nil {
			logrus.Errorf("Unable to open logfile for Thing: %s", filePath)
			return
		} else {
			logrus.Infof("Created logfile for Thing: %s", filePath)
		}
		logger.SetOutput(fileHandle)
		wlog.loggers[thingID] = logger
	}
	logger.WithFields(logrus.Fields{
		"sender":  sender,
		"msgType": msgType,
		"thingID": thingID,
	}).Print(string(payload))
}

// Start connects, subscribe and start the recording
// loggerConfig contains the folder to log to. If it has a relative path it is wrt hub home folder
func (wlog *WostLogger) Start(hubConfig *hubconfig.HubConfig, loggerConfig *WostLoggerConfig) error {
	var err error
	wlog.loggers = make(map[string]*logrus.Logger)
	wlog.loggerConfig = *loggerConfig
	wlog.hubConfig = hubConfig

	// verify the logging folder exists
	if !path.IsAbs(wlog.loggerConfig.LogsFolder) {
		wlog.loggerConfig.LogsFolder = path.Join(hubConfig.Home, wlog.loggerConfig.LogsFolder)
	}
	_, err = os.Stat(wlog.loggerConfig.LogsFolder)
	if err != nil {
		logrus.Errorf("Start: Logging folder does not exist: %s. Setup error: %s", wlog.loggerConfig.LogsFolder, err)
		return err
	}

	wlog.hubConnection = hubclient.NewPluginClient(PluginID, hubConfig)
	wlog.hubConnection.Start(false)

	if wlog.loggerConfig.ThingIDs == nil || len(wlog.loggerConfig.ThingIDs) == 0 {
		// log everything
		wlog.hubConnection.Subscribe("", func(thingID string, msgType string, payload []byte, senderID string) {
			wlog.logToFile(thingID, msgType, payload, senderID)
		})
	} else {
		for _, thingID := range wlog.loggerConfig.ThingIDs {
			wlog.hubConnection.Subscribe(thingID,
				func(evThingID string, msgType string, payload []byte, senderID string) {
					wlog.logToFile(evThingID, msgType, payload, senderID)
				})
		}
	}

	logrus.Infof("Started logger of %d topics", len(wlog.loggerConfig.ThingIDs))
	return err
}

// Stop the logging
func (wlog *WostLogger) Stop() {
	logrus.Info("Stopping logging service")
	if len(wlog.loggerConfig.ThingIDs) == 0 {
		wlog.hubConnection.Unsubscribe("")
	} else {
		for _, thingID := range wlog.loggerConfig.ThingIDs {
			wlog.hubConnection.Unsubscribe(thingID)
		}
	}
	for _, logger := range wlog.loggers {
		logger.Out.(*os.File).Close()
	}
	wlog.loggers = nil
}
