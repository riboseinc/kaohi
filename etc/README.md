
# Kaohi Configuration Syntax

The configuration of `Kaohi` is based on [HCL](https://github.com/hashicorp/hcl) syntax.

## Example configuration

```
log-settings {
	directory = "/var/log/kaohi"
	level = 1                                   // 0 - Silent, 1 - Normal, 2 - Verbose, 3 - Debug
}

listen {
	address = "127.0.0.1:8443"
}

configs "cfg_group1" {                         // defines the log files to be watched
	files = [
		"/var/log/system.log",
		"/var/log/fail.log"
	]
}

commands "cmd_group1" {                         // defines the commands to be watched the output of them
	uid = 501
	interval = 10                               // seconds
	files = [
		"/bin/ls",
		"ifconfig"
	]	
}

rsyslog {
	status = "on"
	listen_address = "*.6800"
	protocol = "tcp"
	local_socket = "/var/log/rsyslog.sock"
}

```
