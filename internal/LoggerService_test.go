package internal_test

import (
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/logger/internal"
	"github.com/wostzone/wostlib-go/pkg/hubclient"
	"github.com/wostzone/wostlib-go/pkg/hubconfig"
	"github.com/wostzone/wostlib-go/pkg/td"
	"github.com/wostzone/wostlib-go/pkg/testenv"
	"github.com/wostzone/wostlib-go/wostapi"
)

var homeFolder string

const zone = "test"
const publisherID = "loggerservice"
const testPluginID = "logger-test"

const loremIpsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor " +
	"incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco " +
	"laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate " +
	"velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, " +
	"sunt in culpa qui officia deserunt mollit anim id est laborum."

var hubConfig *hubconfig.HubConfig
var setupOnce = false
var mosquittoCmd *exec.Cmd

// For running mosquitto in test
// const mosquittoConfigFile = "mosquitto-test.conf"

// TestMain run mosquitto and use the project test folder as the home folder.
// Make sure the certificates exist.
func TestMain(m *testing.M) {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../test")
	mosquittoCmd = testenv.Setup(homeFolder, 0)
	if mosquittoCmd == nil {
		logrus.Fatalf("Unable to setup mosquitto")
	}

	result := m.Run()
	mosquittoCmd.Process.Kill()

	os.Exit(result)
}

// Test starting and stopping of the logger service
func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")

	svc := internal.NewLoggerService()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)
	err := svc.Start(hubConfig)
	assert.NoError(t, err)
	svc.Stop()
}

// Test logging of a published TD
func TestLogTD(t *testing.T) {
	logrus.Infof("--- TestLogTD ---")
	deviceID := "device1"
	thingID1 := td.CreatePublisherThingID(zone, publisherID, deviceID, wostapi.DeviceTypeSensor)
	clientID := "TestLogTD"

	svc := internal.NewLoggerService()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)
	err := svc.Start(hubConfig)

	// create a thing to publish with
	client := hubclient.NewPluginClient(clientID, hubConfig)
	err = client.Start()
	require.Nil(t, err)
	time.Sleep(100 * time.Millisecond)

	tdObj := td.CreateTD(thingID1, wostapi.DeviceTypeSensor)
	client.PublishTD(thingID1, tdObj)

	event := td.CreateThingEvent("event1", nil)
	client.PublishEvent(thingID1, event)

	time.Sleep(1 * time.Second)
	client.Stop()

	assert.NoError(t, err)
	svc.Stop()
}

// Test logging of a specific ID
func TestLogSpecificIDs(t *testing.T) {
	logrus.Infof("--- TestLogSpecificIDs ---")
	thingID1 := "urn:zone1:thing1"
	thingID2 := "urn:zone1:thing2"
	clientID := "TestLogSpecificIDs"

	svc := internal.NewLoggerService()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)
	svc.Config.ThingIDs = []string{thingID2}
	err := svc.Start(hubConfig)
	// create a client to publish with
	client := hubclient.NewPluginClient(clientID, hubConfig)
	err = client.Start()
	require.Nil(t, err)
	time.Sleep(100 * time.Millisecond)

	event := td.CreateThingEvent("event1", nil)
	client.PublishEvent(thingID1, event)

	event = td.CreateThingEvent("event2", nil)
	client.PublishEvent(thingID2, event)

	time.Sleep(1 * time.Second)
	client.Stop()

	assert.NoError(t, err)
	svc.Stop()
}

func TestAltLoggingFolder(t *testing.T) {
	logrus.Infof("--- TestAltLoggingFolder ---")

	svc := internal.NewLoggerService()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)
	svc.Config.LogsFolder = "/tmp"
	err := svc.Start(hubConfig)
	assert.NoError(t, err)

}

func TestBadLoggingFolder(t *testing.T) {
	logrus.Infof("--- TestBadLoggingFolder ---")
	svc := internal.NewLoggerService()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)
	svc.Config.LogsFolder = "/notafolder"
	err := svc.Start(hubConfig)
	assert.Error(t, err)

}
