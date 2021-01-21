/*
 * Copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltha-protos/v3/go/common"
	"github.com/opencord/voltha-protos/v3/go/voltha"
)

const (
	DEFAULT_DEVICE_FORMAT         = "table{{ .Id }}\t{{.Type}}\t{{.Root}}\t{{.ParentId}}\t{{.SerialNumber}}\t{{.AdminState}}\t{{.OperStatus}}\t{{.ConnectStatus}}\t{{.Reason}}"
	DEFAULT_DEVICE_PORTS_FORMAT   = "table{{.PortNo}}\t{{.Label}}\t{{.Type}}\t{{.AdminState}}\t{{.OperStatus}}\t{{.DeviceId}}\t{{.Peers}}"
	DEFAULT_DEVICE_INSPECT_FORMAT = `ID: {{.Id}}
  TYPE:          {{.Type}}
  ROOT:          {{.Root}}
  PARENTID:      {{.ParentId}}
  SERIALNUMBER:  {{.SerialNumber}}
  VLAN:          {{.Vlan}}
  ADMINSTATE:    {{.AdminState}}
  OPERSTATUS:    {{.OperStatus}}
  CONNECTSTATUS: {{.ConnectStatus}}`
	DEFAULT_DEVICE_PM_CONFIG_GET_FORMAT         = "table{{.DefaultFreq}}\t{{.Grouped}}\t{{.FreqOverride}}"
	DEFAULT_DEVICE_PM_CONFIG_METRIC_LIST_FORMAT = "table{{.Name}}\t{{.Type}}\t{{.Enabled}}\t{{.SampleFreq}}"
	DEFAULT_DEVICE_PM_CONFIG_GROUP_LIST_FORMAT  = "table{{.GroupName}}\t{{.Enabled}}\t{{.GroupFreq}}"
	DEFAULT_DEVICE_VALUE_GET_FORMAT             = "table{{.Name}}\t{{.Result}}"
)

type DeviceList struct {
	ListOutputOptions
}

type DeviceCreate struct {
	DeviceType  string `short:"t" required:"true" long:"devicetype" description:"Device type"`
	MACAddress  string `short:"m" long:"macaddress" default:"" description:"MAC Address"`
	IPAddress   string `short:"i" long:"ipaddress" default:"" description:"IP Address"`
	HostAndPort string `short:"H" long:"hostandport" default:"" description:"Host and port"`
}

type DeviceId string

type MetricName string
type GroupName string
type PortNum uint32
type ValueFlag string

type DeviceDelete struct {
	Force bool `long:"force" description:"Delete device forcefully"`
	Args  struct {
		Ids []DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceEnable struct {
	Args struct {
		Ids []DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceDisable struct {
	Args struct {
		Ids []DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceReboot struct {
	Args struct {
		Ids []DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceFlowList struct {
	ListOutputOptions
	FlowIdOptions
	Args struct {
		Id DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePortList struct {
	ListOutputOptions
	Args struct {
		Id DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceInspect struct {
	OutputOptionsJson
	Args struct {
		Id DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePortEnable struct {
	Args struct {
		Id     DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
		PortId PortNum  `positional-arg-name:"PORT_NUMBER" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePortDisable struct {
	Args struct {
		Id     DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
		PortId PortNum  `positional-arg-name:"PORT_NUMBER" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigsGet struct {
	ListOutputOptions
	Args struct {
		Id DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigMetricList struct {
	ListOutputOptions
	Args struct {
		Id DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigGroupList struct {
	ListOutputOptions
	Args struct {
		Id DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigGroupMetricList struct {
	ListOutputOptions
	Args struct {
		Id    DeviceId  `positional-arg-name:"DEVICE_ID" required:"yes"`
		Group GroupName `positional-arg-name:"GROUP_NAME" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigFrequencySet struct {
	OutputOptions
	Args struct {
		Frequency uint32   `positional-arg-name:"FREQUENCY" required:"yes"`
		Id        DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigMetricEnable struct {
	Args struct {
		Id      DeviceId     `positional-arg-name:"DEVICE_ID" required:"yes"`
		Metrics []MetricName `positional-arg-name:"METRIC_NAME" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigMetricDisable struct {
	Args struct {
		Id      DeviceId     `positional-arg-name:"DEVICE_ID" required:"yes"`
		Metrics []MetricName `positional-arg-name:"METRIC_NAME" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigGroupEnable struct {
	Args struct {
		Id     DeviceId    `positional-arg-name:"DEVICE_ID" required:"yes"`
		Groups []GroupName `positional-arg-name:"GROUP_NAME" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigGroupDisable struct {
	Args struct {
		Id     DeviceId    `positional-arg-name:"DEVICE_ID" required:"yes"`
		Groups []GroupName `positional-arg-name:"GROUP_NAME" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceGetExtValue struct {
	ListOutputOptions
	Args struct {
		Id        DeviceId  `positional-arg-name:"DEVICE_ID" required:"yes"`
		Valueflag ValueFlag `positional-arg-name:"VALUE_FLAG" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigSetMaxSkew struct {
	Args struct {
		Id      DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
		MaxSkew uint32   `positional-arg-name:"MAX_SKEW" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceOpts struct {
	List    DeviceList     `command:"list"`
	Create  DeviceCreate   `command:"create"`
	Delete  DeviceDelete   `command:"delete"`
	Enable  DeviceEnable   `command:"enable"`
	Disable DeviceDisable  `command:"disable"`
	Flows   DeviceFlowList `command:"flows"`
	Port    struct {
		List    DevicePortList    `command:"list"`
		Enable  DevicePortEnable  `command:"enable"`
		Disable DevicePortDisable `command:"disable"`
	} `command:"port"`
	Inspect DeviceInspect `command:"inspect"`
	Reboot  DeviceReboot  `command:"reboot"`
	Value   struct {
		Get DeviceGetExtValue `command:"get"`
	} `command:"value"`
	PmConfig struct {
		Get     DevicePmConfigsGet `command:"get"`
		MaxSkew struct {
			Set DevicePmConfigSetMaxSkew `command:"set"`
		} `command:"maxskew"`
		Frequency struct {
			Set DevicePmConfigFrequencySet `command:"set"`
		} `command:"frequency"`
		Metric struct {
			List    DevicePmConfigMetricList    `command:"list"`
			Enable  DevicePmConfigMetricEnable  `command:"enable"`
			Disable DevicePmConfigMetricDisable `command:"disable"`
		} `command:"metric"`
		Group struct {
			List    DevicePmConfigGroupList    `command:"list"`
			Enable  DevicePmConfigGroupEnable  `command:"enable"`
			Disable DevicePmConfigGroupDisable `command:"disable"`
		} `command:"group"`
		GroupMetric struct {
			List DevicePmConfigGroupMetricList `command:"list"`
		} `command:"groupmetric"`
	} `command:"pmconfig"`
}

var deviceOpts = DeviceOpts{}

func RegisterDeviceCommands(parser *flags.Parser) {
	if _, err := parser.AddCommand("device", "device commands", "Commands to query and manipulate VOLTHA devices", &deviceOpts); err != nil {
		Error.Fatalf("Unexpected error while attempting to register device commands : %s", err)
	}
}

func (i *MetricName) Complete(match string) []flags.Completion {
	conn, err := NewConnection()
	if err != nil {
		return nil
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	var deviceId string
found:
	for i := len(os.Args) - 1; i >= 0; i -= 1 {
		switch os.Args[i] {
		case "enable":
			fallthrough
		case "disable":
			if len(os.Args) > i+1 {
				deviceId = os.Args[i+1]
			} else {
				return nil
			}
			break found
		default:
		}
	}

	if len(deviceId) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(deviceId)}

	pmconfigs, err := client.ListDevicePmConfigs(ctx, &id)

	if err != nil {
		return nil
	}

	list := make([]flags.Completion, 0)
	for _, metrics := range pmconfigs.Metrics {
		if strings.HasPrefix(metrics.Name, match) {
			list = append(list, flags.Completion{Item: metrics.Name})
		}
	}

	return list
}

func (i *GroupName) Complete(match string) []flags.Completion {
	conn, err := NewConnection()
	if err != nil {
		return nil
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	var deviceId string
found:
	for i := len(os.Args) - 1; i >= 0; i -= 1 {
		switch os.Args[i] {
		case "list":
			fallthrough
		case "enable":
			fallthrough
		case "disable":
			if len(os.Args) > i+1 {
				deviceId = os.Args[i+1]
			} else {
				return nil
			}
			break found
		default:
		}
	}

	if len(deviceId) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(deviceId)}

	pmconfigs, err := client.ListDevicePmConfigs(ctx, &id)

	if err != nil {
		return nil
	}

	list := make([]flags.Completion, 0)
	for _, group := range pmconfigs.Groups {
		if strings.HasPrefix(group.GroupName, match) {
			list = append(list, flags.Completion{Item: group.GroupName})
		}
	}
	return list
}

func (i *PortNum) Complete(match string) []flags.Completion {
	conn, err := NewConnection()
	if err != nil {
		return nil
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	/*
	 * The command line args when completing for PortNum will be a DeviceId
	 * followed by one or more PortNums. So walk the argument list from the
	 * end and find the first argument that is enable/disable as those are
	 * the subcommands that come before the positional arguments. It would
	 * be nice if this package gave us the list of optional arguments
	 * already parsed.
	 */
	var deviceId string
found:
	for i := len(os.Args) - 1; i >= 0; i -= 1 {
		switch os.Args[i] {
		case "enable":
			fallthrough
		case "disable":
			if len(os.Args) > i+1 {
				deviceId = os.Args[i+1]
			} else {
				return nil
			}
			break found
		default:
		}
	}

	if len(deviceId) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(deviceId)}

	ports, err := client.ListDevicePorts(ctx, &id)
	if err != nil {
		return nil
	}

	list := make([]flags.Completion, 0)
	for _, item := range ports.Items {
		pn := strconv.FormatUint(uint64(item.PortNo), 10)
		if strings.HasPrefix(pn, match) {
			list = append(list, flags.Completion{Item: pn})
		}
	}

	return list
}

func (i *DeviceId) Complete(match string) []flags.Completion {
	conn, err := NewConnection()
	if err != nil {
		return nil
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	devices, err := client.ListDevices(ctx, &empty.Empty{})
	if err != nil {
		return nil
	}

	list := make([]flags.Completion, 0)
	for _, item := range devices.Items {
		if strings.HasPrefix(item.Id, match) {
			list = append(list, flags.Completion{Item: item.Id})
		}
	}

	return list
}

func (options *DeviceList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	devices, err := client.ListDevices(ctx, &empty.Empty{})
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-list", "format", DEFAULT_DEVICE_FORMAT)
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("device-list", "order", "")
	}

	// Make sure json output prints an empty list, not "null"
	if devices.Items == nil {
		devices.Items = make([]*voltha.Device, 0)
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      devices.Items,
	}

	GenerateOutput(&result)
	return nil
}

func (options *DeviceCreate) Execute(args []string) error {

	device := voltha.Device{}
	if options.HostAndPort != "" {
		device.Address = &voltha.Device_HostAndPort{HostAndPort: options.HostAndPort}
	} else if options.IPAddress != "" {
		device.Address = &voltha.Device_Ipv4Address{Ipv4Address: options.IPAddress}
	}
	if options.MACAddress != "" {
		device.MacAddress = strings.ToLower(options.MACAddress)
	}
	if options.DeviceType != "" {
		device.Type = options.DeviceType
	}

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	createdDevice, err := client.CreateDevice(ctx, &device)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", createdDevice.Id)

	return nil
}

func (options *DeviceDelete) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)
	var lastErr error
	for _, i := range options.Args.Ids {
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
		defer cancel()

		id := voltha.ID{Id: string(i)}
		if options.Force {
			_, err = client.ForceDeleteDevice(ctx, &id)
		} else {
			_, err = client.DeleteDevice(ctx, &id)
		}

		if err != nil {
			Error.Printf("Error while deleting '%s': %s\n", i, err)
			lastErr = err
			continue
		}
		fmt.Printf("%s\n", i)
	}

	if lastErr != nil {
		return NoReportErr
	}
	return nil
}

func (options *DeviceEnable) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	var lastErr error
	for _, i := range options.Args.Ids {
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
		defer cancel()

		id := voltha.ID{Id: string(i)}

		_, err := client.EnableDevice(ctx, &id)
		if err != nil {
			Error.Printf("Error while enabling '%s': %s\n", i, err)
			lastErr = err
			continue
		}
		fmt.Printf("%s\n", i)
	}

	if lastErr != nil {
		return NoReportErr
	}
	return nil
}

func (options *DeviceDisable) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	var lastErr error
	for _, i := range options.Args.Ids {
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
		defer cancel()

		id := voltha.ID{Id: string(i)}

		_, err := client.DisableDevice(ctx, &id)
		if err != nil {
			Error.Printf("Error while disabling '%s': %s\n", i, err)
			lastErr = err
			continue
		}
		fmt.Printf("%s\n", i)
	}

	if lastErr != nil {
		return NoReportErr
	}
	return nil
}

func (options *DeviceReboot) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	var lastErr error
	for _, i := range options.Args.Ids {
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
		defer cancel()

		id := voltha.ID{Id: string(i)}

		_, err := client.RebootDevice(ctx, &id)
		if err != nil {
			Error.Printf("Error while rebooting '%s': %s\n", i, err)
			lastErr = err
			continue
		}
		fmt.Printf("%s\n", i)
	}

	if lastErr != nil {
		return NoReportErr
	}
	return nil
}

func (options *DevicePortList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	ports, err := client.ListDevicePorts(ctx, &id)
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-ports", "format", DEFAULT_DEVICE_PORTS_FORMAT)
	}

	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("device-ports", "order", "")
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      ports.Items,
	}

	GenerateOutput(&result)
	return nil
}

func (options *DeviceFlowList) Execute(args []string) error {
	fl := &FlowList{}
	fl.ListOutputOptions = options.ListOutputOptions
	fl.FlowIdOptions = options.FlowIdOptions
	fl.Args.Id = string(options.Args.Id)
	fl.Method = "device-flows"
	return fl.Execute(args)
}

func (options *DeviceInspect) Execute(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("only a single argument 'DEVICE_ID' can be provided")
	}

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	device, err := client.GetDevice(ctx, &id)
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-inspect", "format", DEFAULT_DEVICE_INSPECT_FORMAT)
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      device,
	}
	GenerateOutput(&result)
	return nil
}

/*Device  Port Enable */
func (options *DevicePortEnable) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	port := voltha.Port{DeviceId: string(options.Args.Id), PortNo: uint32(options.Args.PortId)}

	_, err = client.EnablePort(ctx, &port)
	if err != nil {
		Error.Printf("Error enabling port number %v on device Id %s,err=%s\n", options.Args.PortId, options.Args.Id, ErrorToString(err))
		return err
	}

	return nil
}

/*Device  Port Disable */
func (options *DevicePortDisable) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	port := voltha.Port{DeviceId: string(options.Args.Id), PortNo: uint32(options.Args.PortId)}

	_, err = client.DisablePort(ctx, &port)
	if err != nil {
		Error.Printf("Error enabling port number %v on device Id %s,err=%s\n", options.Args.PortId, options.Args.Id, ErrorToString(err))
		return err
	}

	return nil
}

func (options *DevicePmConfigSetMaxSkew) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	pmConfigs.MaxSkew = options.Args.MaxSkew

	_, err = client.UpdateDevicePmConfigs(ctx, pmConfigs)
	if err != nil {
		return err
	}

	return nil
}

func (options *DevicePmConfigsGet) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-pm-configs", "format", DEFAULT_DEVICE_PM_CONFIG_GET_FORMAT)
	}

	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("device-pm-configs", "order", "")
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      pmConfigs,
	}

	GenerateOutput(&result)
	return nil

}

func (options *DevicePmConfigMetricList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	if !pmConfigs.Grouped {
		for _, metric := range pmConfigs.Metrics {
			if metric.SampleFreq == 0 {
				metric.SampleFreq = pmConfigs.DefaultFreq
			}
		}
		outputFormat := CharReplacer.Replace(options.Format)
		if outputFormat == "" {
			outputFormat = GetCommandOptionWithDefault("device-pm-configs", "format", DEFAULT_DEVICE_PM_CONFIG_METRIC_LIST_FORMAT)
		}

		orderBy := options.OrderBy
		if orderBy == "" {
			orderBy = GetCommandOptionWithDefault("device-pm-configs", "order", "")
		}

		result := CommandResult{
			Format:    format.Format(outputFormat),
			Filter:    options.Filter,
			OrderBy:   orderBy,
			OutputAs:  toOutputType(options.OutputAs),
			NameLimit: options.NameLimit,
			Data:      pmConfigs.Metrics,
		}

		GenerateOutput(&result)
		return nil
	} else {
		return fmt.Errorf("Device '%s' does not have Non Grouped Metrics", options.Args.Id)
	}
}

func (options *DevicePmConfigMetricEnable) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	if !pmConfigs.Grouped {
		metrics := make(map[string]struct{})
		for _, metric := range pmConfigs.Metrics {
			metrics[metric.Name] = struct{}{}
		}

		for _, metric := range pmConfigs.Metrics {
			for _, mName := range options.Args.Metrics {
				if _, exist := metrics[string(mName)]; !exist {
					return fmt.Errorf("Metric Name '%s' does not exist", mName)
				}

				if string(mName) == metric.Name && !metric.Enabled {
					metric.Enabled = true
					_, err := client.UpdateDevicePmConfigs(ctx, pmConfigs)
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		return fmt.Errorf("Device '%s' does not have Non Grouped Metrics", options.Args.Id)
	}
	return nil
}

func (options *DevicePmConfigMetricDisable) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	if !pmConfigs.Grouped {
		metrics := make(map[string]struct{})
		for _, metric := range pmConfigs.Metrics {
			metrics[metric.Name] = struct{}{}
		}

		for _, metric := range pmConfigs.Metrics {
			for _, mName := range options.Args.Metrics {
				if _, have := metrics[string(mName)]; !have {
					return fmt.Errorf("Metric Name '%s' does not exist", mName)
				}
				if string(mName) == metric.Name && metric.Enabled {
					metric.Enabled = false
					_, err := client.UpdateDevicePmConfigs(ctx, pmConfigs)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("Metric '%s' cannot be disabled", string(mName))
				}
			}
		}
	} else {
		return fmt.Errorf("Device '%s' does not have Non Grouped Metrics", options.Args.Id)
	}
	return nil
}

func (options *DevicePmConfigGroupEnable) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	if pmConfigs.Grouped {
		groups := make(map[string]struct{})
		for _, group := range pmConfigs.Groups {
			groups[group.GroupName] = struct{}{}
		}
		for _, group := range pmConfigs.Groups {
			for _, gName := range options.Args.Groups {
				if _, have := groups[string(gName)]; !have {
					return fmt.Errorf("Group Name '%s' does not exist", gName)
				}
				if string(gName) == group.GroupName && !group.Enabled {
					group.Enabled = true
					_, err := client.UpdateDevicePmConfigs(ctx, pmConfigs)
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		return fmt.Errorf("Device '%s' does not have Group Metrics", options.Args.Id)
	}
	return nil
}

func (options *DevicePmConfigGroupDisable) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	if pmConfigs.Grouped {
		groups := make(map[string]struct{})
		for _, group := range pmConfigs.Groups {
			groups[group.GroupName] = struct{}{}
		}

		for _, group := range pmConfigs.Groups {
			for _, gName := range options.Args.Groups {
				if _, have := groups[string(gName)]; !have {
					return fmt.Errorf("Group Name '%s' does not exist", gName)
				}

				if string(gName) == group.GroupName && group.Enabled {
					group.Enabled = false
					_, err := client.UpdateDevicePmConfigs(ctx, pmConfigs)
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		return fmt.Errorf("Device '%s' does not have Group Metrics", options.Args.Id)
	}
	return nil
}

func (options *DevicePmConfigGroupList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	if pmConfigs.Grouped {
		for _, group := range pmConfigs.Groups {
			if group.GroupFreq == 0 {
				group.GroupFreq = pmConfigs.DefaultFreq
			}
		}
		outputFormat := CharReplacer.Replace(options.Format)
		if outputFormat == "" {
			outputFormat = GetCommandOptionWithDefault("device-pm-configs", "format", DEFAULT_DEVICE_PM_CONFIG_GROUP_LIST_FORMAT)
		}

		orderBy := options.OrderBy
		if orderBy == "" {
			orderBy = GetCommandOptionWithDefault("device-pm-configs", "order", "")
		}

		result := CommandResult{
			Format:    format.Format(outputFormat),
			Filter:    options.Filter,
			OrderBy:   orderBy,
			OutputAs:  toOutputType(options.OutputAs),
			NameLimit: options.NameLimit,
			Data:      pmConfigs.Groups,
		}

		GenerateOutput(&result)
	} else {
		return fmt.Errorf("Device '%s' does not have Group Metrics", string(options.Args.Id))
	}
	return nil
}

func (options *DevicePmConfigGroupMetricList) Execute(args []string) error {

	var metrics []*voltha.PmConfig
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	for _, groups := range pmConfigs.Groups {

		if string(options.Args.Group) == groups.GroupName {
			for _, metric := range groups.Metrics {
				if metric.SampleFreq == 0 && groups.GroupFreq == 0 {
					metric.SampleFreq = pmConfigs.DefaultFreq
				} else {
					metric.SampleFreq = groups.GroupFreq
				}
			}
			metrics = groups.Metrics
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-pm-configs", "format", DEFAULT_DEVICE_PM_CONFIG_METRIC_LIST_FORMAT)
	}

	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("device-pm-configs", "order", "")
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      metrics,
	}

	GenerateOutput(&result)
	return nil

}

func (options *DevicePmConfigFrequencySet) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	pmConfigs.DefaultFreq = options.Args.Frequency

	_, err = client.UpdateDevicePmConfigs(ctx, pmConfigs)
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-pm-configs", "format", DEFAULT_DEVICE_PM_CONFIG_GET_FORMAT)
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      pmConfigs,
	}

	GenerateOutput(&result)
	return nil

}

type ReturnValueRow struct {
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
}

/*Device  get Onu Distance */
func (options *DeviceGetExtValue) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	valueflag, okay := common.ValueType_Type_value[string(options.Args.Valueflag)]
	if !okay {
		Error.Printf("Unknown valueflag %s\n", options.Args.Valueflag)
	}

	val := voltha.ValueSpecifier{Id: string(options.Args.Id), Value: common.ValueType_Type(valueflag)}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetCurrentStack().Grpc.Timeout)
	defer cancel()

	rv, err := client.GetExtValue(ctx, &val)
	if err != nil {
		Error.Printf("Error getting value on device Id %s,err=%s\n", options.Args.Id, ErrorToString(err))
		return err
	}

	var rows []ReturnValueRow
	for name, num := range common.ValueType_Type_value {
		if num == 0 {
			// EMPTY is not a real value
			continue
		}
		if (rv.Error & uint32(num)) != 0 {
			row := ReturnValueRow{Name: name, Result: "Error"}
			rows = append(rows, row)
		}
		if (rv.Unsupported & uint32(num)) != 0 {
			row := ReturnValueRow{Name: name, Result: "Unsupported"}
			rows = append(rows, row)
		}
		if (rv.Set & uint32(num)) != 0 {
			switch name {
			case "DISTANCE":
				row := ReturnValueRow{Name: name, Result: rv.Distance}
				rows = append(rows, row)
			default:
				row := ReturnValueRow{Name: name, Result: "Unimplemented-in-voltctl"}
				rows = append(rows, row)
			}
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-value-get", "format", DEFAULT_DEVICE_VALUE_GET_FORMAT)
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      rows,
	}
	GenerateOutput(&result)
	return nil
}
