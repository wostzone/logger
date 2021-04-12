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

	"github.com/wostzone/hubapi/api"
	"github.com/wostzone/hubapi/pkg/hubclient"
	"github.com/wostzone/hubapi/pkg/hubconfig"
	"github.com/wostzone/hubapi/pkg/td"
	"github.com/wostzone/hubapi/pkg/testenv"
	"github.com/wostzone/logger/internal"
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
var mcmd *exec.Cmd

// For running mosquitto in test
const mosquittoConfigFile = "mosquitto-test.conf"

// Use the project test folder as the home folder and make sure the certificates exist
func setup() *internal.LoggerService {
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../test")
	mcmd = testenv.Setup(homeFolder, 0)

	svc := internal.NewLoggerService()
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, &svc.Config)
	return svc
}
func teardown() {
	testenv.Teardown(mcmd)
}

// Test starting and stopping of the logger service
func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")

	svc := setup()
	err := svc.Start(hubConfig)
	assert.NoError(t, err)
	svc.Stop()
	// server.Stop()
	teardown()
}

// Test logging of a published TD
func TestLogTD(t *testing.T) {
	logrus.Infof("--- TestLogTD ---")
	deviceID := "device1"
	thingID1 := td.CreatePublisherThingID(zone, publisherID, deviceID, api.DeviceTypeSensor)
	clientID := "TestLogTD"

	svc := setup()
	err := svc.Start(hubConfig)

	// create a thing to publish with
	client := hubclient.NewPluginClient(clientID, hubConfig)
	err = client.Start()
	require.Nil(t, err)
	time.Sleep(100 * time.Millisecond)

	tdObj := td.CreateTD(thingID1, api.DeviceTypeSensor)
	client.PublishTD(thingID1, tdObj)

	event := td.CreateThingEvent("event1", nil)
	client.PublishEvent(thingID1, event)

	time.Sleep(1 * time.Second)
	client.Stop()

	assert.NoError(t, err)
	svc.Stop()
	teardown()
}

// Test logging of a specific ID
func TestLogSpecificIDs(t *testing.T) {
	logrus.Infof("--- TestLogSpecificIDs ---")
	thingID1 := "urn:zone1:thing1"
	thingID2 := "urn:zone1:thing2"
	clientID := "TestLogSpecificIDs"

	svc := setup()
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
	teardown()
}

func TestAltLoggingFolder(t *testing.T) {
	logrus.Infof("--- TestAltLoggingFolder ---")

	svc := setup()
	svc.Config.LogsFolder = "/tmp"
	err := svc.Start(hubConfig)
	assert.NoError(t, err)

	teardown()
}

func TestBadLoggingFolder(t *testing.T) {
	logrus.Infof("--- TestBadLoggingFolder ---")
	svc := setup()
	svc.Config.LogsFolder = "/notafolder"
	err := svc.Start(hubConfig)
	assert.Error(t, err)

	teardown()
}
