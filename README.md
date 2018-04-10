# Kaohi

Kaohi is an advanced log daemon that solves all the problems that currently exist in syslog daemons on Unix.


## Description

Typical Syslog daemon problems (due to the nature of the syslog protocol):

. No authenticity and integrity verification
. Log events are text strings only
. Logging inside containers is cumbersome and can lead to disk space problems


## Kaohi Modules

Kaohi consists of five core modules that provide log collection functionality

### Kaohi Log Event Collector (KLEC)

The native log event model-based collector

### Kaohi File Collector (KFD)

File based log monitoring similar to `tail -f`

### Kaohi Command Collector (KOC)

The Command Collector executes commands at set intervals and logs the output of these commands

### Kaohi Rsyslog Collector (KRC)

The KRC is a lightweight rsyslog replacement suitable for cloud systems or containers

### Kaohi Pipe Collector (KPC)

The KPC is a named pipe log collector best suited to be used in cloud systems or containers


## Kaohi Architecture Overview

![alt text](https://raw.githubusercontent.com/riboseinc/kaohi/master/images/kaohi-modules-and-architecture.png "Kaohi architecture")

