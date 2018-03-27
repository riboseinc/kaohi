
# Kaohi Configuration Syntax

## Example configuration

```
{
	"log_dir": "/var/log/kaohi",
	"log_level": "DEBUG",

	"listen_address": "127.0.0.1:8443",

	"watchers": [
		{
			"name": "test1",
			"file_paths": [
				"/var/log/syslog",
				"/var/log/faillog"
			]
		}
	],

	"commands": [
		{
			"name": "test2",
			"uid": 501,
			"interval": 10,
			"cmds": [
				"/bin/ls",
				"ifconfig"
			]
		}
	],

	"rsyslog": {
		"listen_address": "*:6880",
		"protocol": "tcp",
		"local_socket": "/var/run/kaohi.sock",
	}
}
```

## Description for example configuration

- `log_dir`: The logging directory of `kaohi` daemon. The default is `/var/log/kaohi`.
- `log_level`: The logging level of `kaohi` daemon. It may have `SILENT`, `NORMAL`, `VERBOSE`, `DEBUG`. The default is `NORMAL`.
- `listen_address`: The listening address for incoming local connections from `kaohi_console`. The default is `127.0.0.1:8443`.
- `watchers`: The `array` of log files to be monitored.
    `name`: The `watcher` item name.
    `file_paths`: The log file paths which are belong to given `watcher` item.

- `commands`: The `array` of commands to be captured.<br />
    `name`: The `command` item name.<br />
    `uid`: The `UID` to run command. The `default` is 0(i.e. root)<br />
    `interval`: The `interval` between continuous executions of `command`. The `default` is `1hour`<br />
    `cmds`: The `array` of the commands to be executed.

- `rsyslog`: The `rsyslog` information to get remote sysloggings.<br />
    `listen_address`: The address to be listen for incoming syslogging connections. The address format is `IPv4Addr:port` or `IPv6Addr:port`<br />
    `protocol`: The protocol type for listening. The types may be `tcp`, `udp`, `tcp6`, `udp6`<br />
    `local_socket`: The local socket path for listening. This option is exclusive with `listen_address` option.
