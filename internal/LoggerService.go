package internal

import (
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wostzone/hubapi/api"
	hubapi "github.com/wostzone/hubapi/api"
	"github.com/wostzone/hubapi/pkg/hubclient"
	"github.com/wostzone/hubapi/pkg/hubconfig"
	"github.com/wostzone/hubapi/pkg/td"
)

// PluginID is the default ID of the WoST Logger plugin
const PluginID = "logger"

// WostLoggerConfig with logger plugin configuration
// map of topic -> file
type WostLoggerConfig struct {
	ClientID   string   `yaml:"clientID"`   // custom unique client ID of logger instance
	PublishTD  bool     `yaml:"publishTD"`  // publish the TD of this service
	LogsFolder string   `yaml:"logsFolder"` // folder to use for logging
	ThingIDs   []string `yaml:"thingIDs"`   // thing IDs to log
}

// LoggerService is a hub plugin for recording messages to the hub
// By default it logs messages by ThingID, eg each Thing has a log file
type LoggerService struct {
	Config        WostLoggerConfig
	hubConfig     *hubconfig.HubConfig
	hubConnection hubapi.IHubClient
	loggers       map[string]*os.File // map of thing ID to logfile
}

// handleMessage receives and records a topic message
func (wlog *LoggerService) logToFile(thingID string, msgType string, payload []byte, sender string) {
	logrus.Infof("Received message of type '%s' about Thing %s", msgType, thingID)
	// var err error

	if wlog.loggers == nil {
		logrus.Errorf("logToFile called after logger has stopped")
		return
	}

	logger := wlog.loggers[thingID]
	if logger == nil {
		logsFolder := wlog.Config.LogsFolder
		filePath := path.Join(logsFolder, thingID+".log")

		// 	TimestampFormat: "2006-01-02T15:04:05.000-0700",
		fileHandle, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0640)
		if err != nil {
			logrus.Errorf("Unable to open logfile for Thing: %s", filePath)
			return
		} else {
			logrus.Infof("Created logfile for Thing: %s", filePath)
		}
		wlog.loggers[thingID] = fileHandle
		logger = fileHandle
	}
	parsedMsg := make(map[string]interface{})
	logMsg := make(map[string]interface{})
	logMsg["receivedAt"] = time.Now().Format("2006-01-02T15:04:05.000-0700")
	logMsg["sender"] = ""
	logMsg["payload"] = parsedMsg
	logMsg["thingID"] = thingID
	logMsg["msgType"] = msgType
	json.Unmarshal(payload, &parsedMsg)
	pretty, _ := json.MarshalIndent(logMsg, " ", "  ")
	prettyStr := string(pretty) + ",\n"
	logger.WriteString(prettyStr)
}

// PublishServiceTD publishes the Thing Description of the logger service
func (wlog *LoggerService) PublishServiceTD() {
	if !wlog.Config.PublishTD {
		return
	}
	deviceType := api.DeviceTypeService
	thingID := td.CreatePublisherThingID(wlog.hubConfig.Zone, "hub", wlog.Config.ClientID, deviceType)
	logrus.Infof("Publishing this service TD %s", thingID)
	thingTD := td.CreateTD(thingID, deviceType)
	// Include the logging folder as a property
	prop := td.CreateProperty("Logging Folder", "Directory where to store the log files", api.PropertyTypeAttr)
	td.SetPropertyDataTypeString(prop, 0, 0)
	//
	td.AddTDProperty(thingTD, "logsFolder", prop)
	wlog.hubConnection.PublishTD(thingID, thingTD)
	td.SetThingDescription(thingTD, "Simple Hub message logging", "This service logs hub messages to file")
}

// Start connects, subscribe and start the recording
func (wlog *LoggerService) Start(hubConfig *hubconfig.HubConfig) error {
	var err error
	// wlog.loggers = make(map[string]*logrus.Logger)
	wlog.loggers = make(map[string]*os.File)
	wlog.hubConfig = hubConfig

	// verify the logging folder exists
	if wlog.Config.LogsFolder == "" {
		// default location is hubConfig log folder
		hubLogFolder := path.Dir(hubConfig.Logging.LogFile)
		wlog.Config.LogsFolder = hubLogFolder
	} else if !path.IsAbs(wlog.Config.LogsFolder) {
		wlog.Config.LogsFolder = path.Join(hubConfig.Home, wlog.Config.LogsFolder)
	}
	_, err = os.Stat(wlog.Config.LogsFolder)
	if err != nil {
		logrus.Errorf("Start: Logging folder does not exist: %s. Setup error: %s", wlog.Config.LogsFolder, err)
		return err
	}

	wlog.hubConnection = hubclient.NewPluginClient(wlog.Config.ClientID, hubConfig)
	wlog.hubConnection.Start(false)

	if wlog.Config.ThingIDs == nil || len(wlog.Config.ThingIDs) == 0 {
		// log everything
		wlog.hubConnection.Subscribe("", func(thingID string, msgType string, payload []byte, senderID string) {
			wlog.logToFile(thingID, msgType, payload, senderID)
		})
	} else {
		for _, thingID := range wlog.Config.ThingIDs {
			wlog.hubConnection.Subscribe(thingID,
				func(evThingID string, msgType string, payload []byte, senderID string) {
					wlog.logToFile(evThingID, msgType, payload, senderID)
				})
		}
	}

	// publish the logger service thing
	wlog.PublishServiceTD()

	logrus.Infof("Started logger of %d topics", len(wlog.Config.ThingIDs))
	return err
}

// Stop the logging
func (wlog *LoggerService) Stop() {
	logrus.Info("Stopping logging service")
	if len(wlog.Config.ThingIDs) == 0 {
		wlog.hubConnection.Unsubscribe("")
	} else {
		for _, thingID := range wlog.Config.ThingIDs {
			wlog.hubConnection.Unsubscribe(thingID)
		}
	}
	for _, logger := range wlog.loggers {
		// logger.Out.(*os.File).Close()
		logger.Close()
	}
	wlog.loggers = nil
	wlog.hubConnection.Stop()
}

// NewLoggerService returns a new instance of the logger service
func NewLoggerService() *LoggerService {
	svc := &LoggerService{
		Config: WostLoggerConfig{
			ClientID:   PluginID,
			PublishTD:  false,
			LogsFolder: "",
		},
	}
	return svc
}
