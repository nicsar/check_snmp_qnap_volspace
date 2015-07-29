# check_snmp_qnap_volspace

This is a Nagios plugin written in *Go* to check status and space usage of Qnap volumes via *SNMP*.

## Usage
From command line:

	nagios:~$./check_snmp_qnap_volspace -h
	Usage: check_snmp_qnap_volspace -H <host> [-C <snmp_community>] [-p <port>] [-t <timeout>]
  	-C="public": community name for the host's SNMP agent
  	-H="": name or IP address of host to check
  	-c=90: percent of space volume used to generate CRITICAL state
  	-f=false: perfparse compatible output
  	-p="161": SNMP port
  	-t=10: timeout for SNMP in seconds
  	-w=80: percent of space volume used to generate WARNING state

Example:

	nagios:~$ ./check_snmp_qnap_volspace -H 10.1.5.10 -C gpublic -w 80 -c 90
	OK: volumes free space Ok - volumes status Ok
	
	nagios:~$ ./check_snmp_qnap_volspace -H 10.1.5.10 -C gpublic -w 30 -c 90 
	WARNING: [Volume Volume-2, Pool 2] above warning threshold

## Nagios integration
Define a command like this:

	define command {
	 command_name  check_qnap_snmp_volspace
	 command_line  $USER1$/check_snmp_qnap_volspace -H $HOSTADDRESS$ -C $ARG1$ -f -w $ARG2$ -c $ARG3$
	 register      1
	}

### Performance data
Collecting of performance data can be done by using the `-f` flag.
Performance data can be graphed using `pnp4nagios`:

	nagios:~# cd /usr/local/pnp4nagios/share/templates
	nagios:templates# ln -sv ../templates.dist/check_disk.php check_qnap_snmp_volspace.php
	
## Compilation and installation
Before compile you need to get [nagutils](https://github.com/nicsar/nagutils).
To compile `cd` to `check_snmp_qnap_volspace` directory, then copy the executable to *Nagios* plugin directory:

	nagios:~$ go build
	nagios:~$ sudo cp check_snmp_qnap_volspace /usr/local/nagios/libexec/

### Cross compiling
To compile for *Linux* 32 bit:
	
	~$ GOARCH=386 GOOS=linux go build

To compile fro *FreeBSD* 32 bit:

	~$ GOARCH=386 GOOS=freebsd go build
