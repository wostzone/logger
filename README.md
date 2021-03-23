# WoST Hub Logger

Simple logger of messages on the hub message bus, intended for testing of plugins and things.


## Objective

Facilitate the development of plugins and Things by logging messages from the hub message bus.


## Status 

The status of this plugin is Alpha.

Basic logging of Thing messages to file is functional.


## Audience

This project is aimed at software developers, system implementors and people with a keen interest in the Web of Things. 

## Summary

Things and Hub plugins publish information on Thing TD's, events and actions over the message bus. This plugin writes those messages to file. Each thing has its own file to enable testing the output of each Thing.

## Build and Installation

### System Requirements

This plugin runs as a plugin of the WoST hub. It has no additional requirements other than a working hub message bus.


### Manual Installation

See the Hub README on plugin installation.
In short: copy this plugin to the Hub bin folder and the logger.yaml config file to the Hub's config folder. Add the logger module to the list of plugins to launch on startup and restart the Hub.


### Build From Source

Build with:
```
make all
```

The plugin can be found in dist/bin for 64bit intel or amd processors, or dist/arm for 64 bit ARM processors. Copy this to the hub bin or arm directory.
An example configuration file is provided in config/logger.yaml. Copy this to the hub config directory.

## Usage

When installed as a Hub plugin this plugin is launched automatically on startup of the Hub. It generates log files in the configured logging folder with the pluginID as the filename. Eg:  {apphome}/logs/{thingID}.log

Instead of logging all Things, the logger.yaml configuration file can be configured with the ID's of the Things to log. The rest will be ignored.
