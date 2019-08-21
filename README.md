[![License](https://img.shields.io/github/license/ciena/voltctl.svg)]() [![Project Status: WIP – Initial development is in progress, but there has not yet been a stable, usable release suitable for the public.](https://www.repostatus.org/badges/latest/wip.svg)](https://www.repostatus.org/#wip) [![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://gitHub.com/ciena/voltctl/graphs/commit-activity) [![made-with-python](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://www.golang.org/) [![GitHub contributors](https://img.shields.io/github/contributors/ciena/voltctl.svg)](https://gitHub.com/ciena/voltctl/graphs/contributors/) [![GitHub issues](https://img.shields.io/github/issues/ciena/voltctl.svg)](https://gitHub.com/ciena/voltctl/issues/) [![GitHub issues-closed](https://img.shields.io/github/issues-closed/ciena/voltctl.svg)](https://gitHub.com/ciena/voltctl/issues?q=is%3Aissue+is%3Aclosed) [![Last Commit](https://img.shields.io/github/last-commit/ciena/voltctl.svg)](https://github.com/ciena/voltctl/commits/master)

# `voltctl` - A command line tools to access VOLTHA
In today's VOLTHA implementation the operator interacts with VOLTHA via a CLI
that is accessed by `SSH`ing to the VOLTHA tool. While is can be convenent as
it requires not external tool to be installed it is an abirtation in terms of
industry trends for tools such as `docker` and `kubernetes`, which are both
tools that VOLTHA leverages.

This repository contains a tool that attempts to provide a use model for
VOLTHA that is similar to that of `docker` and `kubernetes` in that a simple
control application is provided to invoke various funcs and the output can
be displayed as a customized/filtered table or as `JSON`.

## Build / Install
To install the `voltctl` command you can use the following:
```bash
mkdir myworkdir
cd myworkdir
export GOPATH=$(pwd)
git clone http://github.com/ciena/voltctl src/github.com/ciena/voltctl
cd src/github.com/ciena/voltctl
make build
cp ./voltctl <to any place you want in your path>
```

`voltctl` has only been tested with `go` version 1.12.x.

## Shell Completion
`voltctl` supports shell completion for the `bash` shell. To enable
shell Completion you can use the following command on *most* \*nix based system.
```bash
source <(voltctl completion bash)
```

If this does not work on your system, as is the case with the standard
bash shell on MacOS, then you can try the following command:
```bash
source /dev/stdin <<<"$(voltctl completion bash)"
```

If you which to make `bash` shell completion automatic when you login to
your account you can append the output of `voltctl completion bash` to
your `$HOME/.bashrc`:
```bash
voltctl completion base >> $HOME/.bashrc
```

## Configuration
Currently the configuration only supports the specification of the VOLTHA
server. There is a sample configuration file name `voltctl.config`. You can
copy this to `~/.volt/config` and modify the server parameter to your
environment. Alternatively you can specify the server on the command line as
well, `voltctl -server host:port ...`.

## Commands
Currently only two commands are working
- `voltctl adapter list` - displays the installed adapters
- `voltctl device list` - displays the devices in the system
- `voltctl device create [-t type] [-i ipv4] [-m mac] [-H host_and_port]` -
  create or pre-provision a device
- `voltctl delete DEVICE_ID [DEVICE_ID...]` - delete one or more devices
- `voltctl enable DEVICE_ID [DEVICE_ID...]` - enable one or more devices
- `voltctl disable DEVICE_ID [DEVICE_ID...]` - disable one or more devices
- `voltctl version` - display the client and server version

## Output Format
Each command has a default output table format. This can be overriden from
the command line using the `voltctl --format=...` option. The specification
of the format is roughly equivalent to the `docker` or `kubectl` command. If
the prefix `table` is specified a table with headers will be displayed, else
each line will be output as specified.

The output of a command may also be written as `JSON` or `YAML` by using the
`--outputas` or `-o` command line option. Valid values for this options are
`table`, `json`, or `yaml`.

### Examples
```
voltctl adapter list
ID                      VENDOR                       VERSION
acme                    Acme Inc.                    0.1
adtran_olt              ADTRAN, Inc.                 1.36
adtran_onu              ADTRAN, Inc.                 1.25
asfvolt16_olt           Edgecore                     0.98
brcm_openomci_onu       Voltha project               0.50
broadcom_onu            Voltha project               0.46
cig_olt                 CIG Tech                     0.11
cig_openomci_onu        CIG Tech                     0.10
dpoe_onu                Sumitomo Electric, Inc.      0.1
maple_olt               Voltha project               0.4
microsemi_olt           Microsemi / Celestica        0.2
openolt                 OLT white box vendor         0.1
pmcs_onu                PMCS                         0.1
ponsim_olt              Voltha project               0.4
ponsim_onu              Voltha project               0.4
simulated_olt           Voltha project               0.1
simulated_onu           Voltha project               0.1
tellabs_olt             Tellabs Inc.                 0.1
tellabs_openomci_onu    Tellabs Inc.                 0.1
tibit_olt               Tibit Communications Inc.    0.1
tibit_onu               Tibit Communications Inc.    0.1
tlgs_onu                TLGS                         0.1
```

```
voltctl device list
ID                  TYPE          ROOT     PARENTID            SERIALNUMBER      VLAN    ADMINSTATE    OPERSTATUS    CONNECTSTATUS
00015bbbfdb3c068    ponsim_olt    true     0001aabbccddeeff    10.1.4.4:50060    0       ENABLED       ACTIVE        REACHABLE
0001552615104a2c    ponsim_onu    false    00015bbbfdb3c068    PSMO12345678      128     ENABLED       ACTIVE        REACHABLE
```

```
voltctl device list --format 'table{{.Id}}\t{{.SerialNumber}}\t{{.ConnectStatus}}'
ID                  SERIALNUMBER      CONNECTSTATUS
00015bbbfdb3c068    10.1.4.4:50060    REACHABLE
0001552615104a2c    PSMO12345678      REACHABLE
```

```
voltctl --server voltha:50555 device list --format '{{.Id}},{{.SerialNumber}},{{.ConnectStatus}}'
00015bbbfdb3c068,10.1.4.4:50060,REACHABLE
0001552615104a2c,PSMO12345678,REACHABLE
````

```
voltctl device list --outputas json
[{"id":"00015bbbfdb3c068","type":"ponsim_olt","root":true,"parent_id":"0001aabbccddeeff","vendor":"ponsim","model":"n/a","serial_number":"10.1.4.4:50060","adapter":"ponsim_olt","Address":{"HostAndPort":"10.1.4.4:50060"},"admin_state":3,"oper_status":4,"connect_status":2},{"id":"0001552615104a2c","type":"ponsim_onu","parent_id":"00015bbbfdb3c068","parent_port_no":1,"vendor":"ponsim","model":"n/a","serial_number":"PSMO12345678","vlan":128,"Address":null,"proxy_address":{"device_id":"00015bbbfdb3c068","channel_id":128},"admin_state":3,"oper_status":4,"connect_status":2}]
```

WIP - just barely started
