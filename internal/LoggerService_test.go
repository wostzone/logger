package internal_test

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wostzone/hubapi/pkg/hubclient"
	"github.com/wostzone/hubapi/pkg/hubconfig"
	"github.com/wostzone/hubapi/pkg/td"
	"github.com/wostzone/logger/internal"
)

var homeFolder string

const testPluginID = "logger-test"
const loremIpsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor " +
	"incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco " +
	"laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate " +
	"velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, " +
	"sunt in culpa qui officia deserunt mollit anim id est laborum."

var loggerConfig *internal.WostLoggerConfig = &internal.WostLoggerConfig{} // use defaults
var hubConfig *hubconfig.HubConfig
var setupOnce = false

// --- NOTE: THIS REQUIRES A RUNNING HUB ---

// Use the project test folder as the home folder
func setup() {
	if setupOnce {
		return
	}
	setupOnce = true
	cwd, _ := os.Getwd()
	homeFolder = path.Join(cwd, "../test")
	// homeFolder = path.Join(cwd, "../../hub/dist")
	loggerConfig = &internal.WostLoggerConfig{}
	os.Args = append(os.Args[0:1], strings.Split("", " ")...)
	hubConfig, _ = hubconfig.LoadPluginConfig(homeFolder, testPluginID, loggerConfig)
}
func teardown() {
}

// Test starting and stopping of the logger service
func TestStartStop(t *testing.T) {
	logrus.Infof("--- TestStartStop ---")
	setup()

	svc := internal.WostLogger{}
	err := svc.Start(hubConfig, loggerConfig)
	assert.NoError(t, err)
	svc.Stop()
	// server.Stop()
	teardown()
}

// Test logging of a published TD
func TestLogTD(t *testing.T) {
	logrus.Infof("--- TestLogTD ---")
	thingID1 := "urn:zone1:thing1:hello"
	clientID := "TestLogTD"
	setup()

	svc := internal.WostLogger{}
	err := svc.Start(hubConfig, loggerConfig)
	// create a thing to publish with
	client := hubclient.NewPluginClient(clientID, hubConfig)
	err = client.Start(false)
	require.Nil(t, err)
	time.Sleep(100 * time.Millisecond)

	tdObj := td.CreateTD(thingID1)
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
	setup()

	svc := internal.WostLogger{}
	loggerConfig.ThingIDs = []string{thingID2}
	err := svc.Start(hubConfig, loggerConfig)
	// create a client to publish with
	client := hubclient.NewPluginClient(clientID, hubConfig)
	err = client.Start(false)
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
	setup()

	svc := internal.WostLogger{}
	loggerConfig.LogsFolder = "/tmp"
	err := svc.Start(hubConfig, loggerConfig)
	assert.Error(t, err)

	teardown()
}

func TestBadLoggingFolder(t *testing.T) {
	logrus.Infof("--- TestBadLoggingFolder ---")
	setup()

	svc := internal.WostLogger{}
	loggerConfig.LogsFolder = "/notafolder"
	err := svc.Start(hubConfig, loggerConfig)
	assert.Error(t, err)

	teardown()
}
