# Libraries for Martin's home automation

This repository contains some libraries which help to implement brokers or custom logic for Martin's home automation.

## MQTT helper

The package `mqtthelper` contains some functions to simplify the connection setup to MQTT.
It also contains helper functions for subscribing to topics and publishing messages.

Additionally there are functions to load a configuration file which can be used in a broker.

There are also some message formats defined which can help implementing a broker.

## Media Center

Since media centers can consist of different software I introduced the package `mediacenter` to define the format of some messages.
Brokers which talk to media center software should use those predefined messages.
This way a custom logic can implement a router which can publish messages for different media center software transparently.

## Service check

The package `servicecheck` provides functions to check the availability of a service.
In this context a service means a port on a host.
Technically the functions try to open a connection to the given address to verify it's availability.

## State store

The struct `StateStore` from the package `statestore` can be used to store various states in a key-value-store.
You can also fetch the states and especially wait for specific state.

This can be used in your custom logic to store some states and synchronize the parallel execution.

# What is Martin's home automation

Martin's home automation is just kind of a container for various software to build my home automation.
The automation is mainly based on [MQTT](http://mqtt.org/), but doesn't define any restrictions or guidelines.

## Brokers

Brokers build the bridge between devices or other external services and MQTT.
A broker can implement one or both of the following behaviours.

- Subscribe to topics and pass the message to the device/service
- React on changes of the device/service and publish messages to topics

The broker is always responsible of the format of the messages.
The used topics should be configurable by the user.

Brokers never implement logic to automate things or to talk to other brokers directly.

### List of available brokers

- SIS-PM
- SystemCtl
- CEC
- LGTV
- MPD
- Kodi

## Custom logic

Your very own automation is defined as custom logic and is implemented in a separate software.
You can subscribe to various topics and publish messages to topics based on your logic.
This software can be implemented in any language which is able to talk to MQTT, but the libraries can only be used in Go.

[My personal custom logic](https://github.com/mjanser/home-automation) can be used as an example.
