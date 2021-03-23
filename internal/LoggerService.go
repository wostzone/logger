package internal

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

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
	fileHandles   map[string]*os.File // map of thing ID to file handle
}

// handleMessage receives and records a topic message
// FIXME: THIS IMPLEMENTS KNOWLEDGE OF THE URL SCHEMA.
func (wlog *WostLogger) logToFile(topic string, msgType string, payload []byte, sender string) {
	logrus.Infof("Received message on topic %s: %s", topic, payload)
	thingID := ""
	parts := strings.Split(topic, "/")
	if len(parts) < 2 {
		return
	}
	thingID = parts[1] // FIXME: This doesnt belong here

	fileHandle := wlog.fileHandles[thingID]
	if fileHandle == nil {
		logsFolder := path.Dir(wlog.loggerConfig.LogsFolder)
		filePath := path.Join(logsFolder, thingID)
		// FIXME: remove invalid tokens from ID
		fileHandle, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)
		if err != nil {
			logrus.Warnf("Unable to open logfile for Thing: %s", filePath)
		}
		wlog.fileHandles[thingID] = fileHandle
	}
	timeStamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
	// timeStamp := time.Now().Format(time.RFC3339Nano)
	maxLen := len(payload)
	if maxLen > 40 {
		maxLen = 40
	}
	line := fmt.Sprintf("[%s] %s %s: %s", timeStamp, sender, topic, payload[:maxLen])
	n, err := fileHandle.WriteString(line + "\n")
	_ = n
	if err != nil {
		logrus.Errorf("Unable to record topic '%s': %s", topic, err)
	}
}

// handleMessage receives and records a topic message
// func (wlog *WostLogger) logToFile(topic string, message []byte, sender string, filename string) {
// 	logrus.Infof("Received message on topic %s: %s", topic, message)

// 	fileHandle := wlog.fileHandles[filename]
// 	if fileHandle != nil {
// 		sender := ""
// 		timeStamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")
// 		// timeStamp := time.Now().Format(time.RFC3339Nano)
// 		maxLen := len(message)
// 		if maxLen > 40 {
// 			maxLen = 40
// 		}
// 		line := fmt.Sprintf("[%s] %s %s: %s", timeStamp, sender, topic, message[:maxLen])
// 		n, err := fileHandle.WriteString(line + "\n")
// 		_ = n
// 		if err != nil {
// 			logrus.Errorf("Unable to record topic '%s': %s", topic, err)
// 		}
// 	}
// }

// AddTopic adds a topic subscription for logging to file
// Return error if logfile can't be opened
// func (wlog *WostLogger) AddTopic(topic string, filename string) error {
// 	logToFile := filename
// 	wlog.hubConnection.Subscribe(topic, func(topic string, message []byte, senderID string) {
// 		wlog.logToFile(topic, message, senderID, logToFile)
// 	})

// 	fileHandle := wlog.fileHandles[filename]
// 	if fileHandle != nil {
// 		// file handle already exists, we're done here
// 		return nil
// 	}
// 	logsFolder := path.Dir(wlog.loggerConfig.LogFolder)
// 	filePath := path.Join(logsFolder, filename)
// 	fileHandle, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)

// 	if err != nil {
// 		logrus.Errorf("Unable to open file '%s' for writing: %s. Topic '%s' ignored", filename, err, topic)
// 		return err
// 	}
// 	wlog.fileHandles[filename] = fileHandle
// 	return nil
// }

// Start connects, subscribe and start the recording
// loggerConfig contains the folder to log to. If it has a relative path it is wrt hub home folder
func (wlog *WostLogger) Start(hubConfig *hubconfig.HubConfig, loggerConfig *WostLoggerConfig) error {
	var err error
	wlog.fileHandles = make(map[string]*os.File)
	wlog.loggerConfig = *loggerConfig
	wlog.hubConfig = hubConfig

	// verify the logging folder exists
	if !path.IsAbs(loggerConfig.LogsFolder) {
		loggerConfig.LogsFolder = path.Join(hubConfig.Home, loggerConfig.LogsFolder)
	}
	_, err = os.Stat(loggerConfig.LogsFolder)
	if err != nil {
		logrus.Errorf("Start: Logging folder does not exist: %s. Setup error: %s", loggerConfig.LogsFolder, err)
		return err
	}

	wlog.hubConnection = hubclient.NewPluginClient(PluginID, hubConfig)

	if loggerConfig.ThingIDs == nil || len(loggerConfig.ThingIDs) == 0 {
		// log everything
		wlog.hubConnection.Subscribe("", func(topic string, msgType string, payload []byte, senderID string) {
			wlog.logToFile(topic, msgType, payload, senderID)
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
	if len(wlog.loggerConfig.ThingIDs) != 0 {
		wlog.hubConnection.Unsubscribe("")
	} else {
		for _, thingID := range wlog.loggerConfig.ThingIDs {
			wlog.hubConnection.Unsubscribe(thingID)
		}
	}
	for _, fileHandle := range wlog.fileHandles {
		fileHandle.Close()
	}
	wlog.fileHandles = make(map[string]*os.File)

}
