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
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltha-protos/v5/go/common"
	"github.com/opencord/voltha-protos/v5/go/extension"
	"github.com/opencord/voltha-protos/v5/go/voltha"
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
	DEFAULT_DEVICE_IMAGE_LIST_GET_FORMAT        = "table{{.Name}}\t{{.Url}}\t{{.Crc}}\t{{.DownloadState}}\t{{.ImageVersion}}\t{{.LocalDir}}\t{{.ImageState}}\t{{.FileSize}}"
	ONU_IMAGE_LIST_FORMAT                       = "table{{.Version}}\t{{.IsCommited}}\t{{.IsActive}}\t{{.IsValid}}\t{{.ProductCode}}\t{{.Hash}}"
	ONU_IMAGE_STATUS_FORMAT                     = "table{{.DeviceId}}\t{{.ImageState.Version}}\t{{.ImageState.DownloadState}}\t{{.ImageState.Reason}}\t{{.ImageState.ImageState}}\t"
	DEFAULT_DEVICE_GET_PORT_STATUS_FORMAT       = `
  TXBYTES:		{{.TxBytes}}
  TXPACKETS:		{{.TxPackets}}
  TXERRPACKETS:		{{.TxErrorPackets}}
  TXBCASTPACKETS:	{{.TxBcastPackets}}
  TXUCASTPACKETS:	{{.TxUcastPackets}}
  TXMCASTPACKETS:	{{.TxMcastPackets}}
  RXBYTES:		{{.RxBytes}}
  RXPACKETS:		{{.RxPackets}}
  RXERRPACKETS:		{{.RxErrorPackets}}
  RXBCASTPACKETS:	{{.RxBcastPackets}}
  RXUCASTPACKETS:	{{.RxUcastPackets}}
  RXMCASTPACKETS:	{{.RxMcastPackets}}`
	DEFAULT_DEVICE_GET_UNI_STATUS_FORMAT = `
  ADMIN_STATE:          {{.AdmState}}
  OPERATIONAL_STATE:    {{.OperState}}
  CONFIG_IND:           {{.ConfigInd}}`
	DEFAULT_ONU_PON_OPTICAL_INFO_STATUS_FORMAT = `
  POWER_FEED_VOLTAGE__VOLTS:      {{.PowerFeedVoltage}}
  RECEIVED_OPTICAL_POWER__dBm:    {{.ReceivedOpticalPower}}
  MEAN_OPTICAL_LAUNCH_POWER__dBm: {{.MeanOpticalLaunchPower}}
  LASER_BIAS_CURRENT__mA:         {{.LaserBiasCurrent}}
  TEMPERATURE__Celsius:           {{.Temperature}}`
	DEFAULT_RX_POWER_STATUS_FORMAT = `
	INTF_ID: {{.IntfId}}
	ONU_ID: {{.OnuId}}
	STATUS: {{.Status}}
	FAIL_REASON: {{.FailReason}}
	RX_POWER : {{.RxPower}}`
	DEFAULT_ETHERNET_FRAME_EXTENDED_PM_COUNTERS_FORMAT = `Upstream_Drop_Events:	        {{.UDropEvents}}
Upstream_Octets:	        {{.UOctets}}
UFrames:	                {{.UFrames}}
UBroadcastFrames:	        {{.UBroadcastFrames}}
UMulticastFrames:	        {{.UMulticastFrames}}
UCrcErroredFrames:	        {{.UCrcErroredFrames}}
UUndersizeFrames:	        {{.UUndersizeFrames}}
UOversizeFrames:	        {{.UOversizeFrames}}
UFrames_64Octets:	        {{.UFrames_64Octets}}
UFrames_65To_127Octets:	        {{.UFrames_65To_127Octets}}
UFrames_128To_255Octets:	{{.UFrames_128To_255Octets}}
UFrames_256To_511Octets:	{{.UFrames_256To_511Octets}}
UFrames_512To_1023Octets:	{{.UFrames_512To_1023Octets}}
UFrames_1024To_1518Octets:	{{.UFrames_1024To_1518Octets}}
DDropEvents:	                {{.DDropEvents}}
DOctets:	                {{.DOctets}}
DFrames:	                {{.DFrames}}
DBroadcastFrames:	        {{.DBroadcastFrames}}
DMulticastFrames:	        {{.DMulticastFrames}}
DCrcErroredFrames:	        {{.DCrcErroredFrames}}
DUndersizeFrames:	        {{.DUndersizeFrames}}
DOversizeFrames:	        {{.DOversizeFrames}}
DFrames_64Octets:	        {{.DFrames_64Octets}}
DFrames_65To_127Octets:	        {{.DFrames_65To_127Octets}}
DFrames_128To_255Octets:	{{.DFrames_128To_255Octets}}
DFrames_256To_511Octets:	{{.DFrames_256To_511Octets}}
DFrames_512To_1023Octets:	{{.DFrames_512To_1023Octets}}
DFrames_1024To_1518Octets:	{{.DFrames_1024To_1518Octets}}`
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

type DeviceFlowGroupList struct {
	ListOutputOptions
	GroupListOptions
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
		Id       DeviceId      `positional-arg-name:"DEVICE_ID" required:"yes"`
		Interval time.Duration `positional-arg-name:"INTERVAL" required:"yes"`
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
		Id    DeviceId  `positional-arg-name:"DEVICE_ID" required:"yes"`
		Group GroupName `positional-arg-name:"GROUP_NAME" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigGroupDisable struct {
	Args struct {
		Id    DeviceId  `positional-arg-name:"DEVICE_ID" required:"yes"`
		Group GroupName `positional-arg-name:"GROUP_NAME" required:"yes"`
	} `positional-args:"yes"`
}

type DevicePmConfigGroupFrequencySet struct {
	OutputOptions
	Args struct {
		Id       DeviceId      `positional-arg-name:"DEVICE_ID" required:"yes"`
		Group    GroupName     `positional-arg-name:"GROUP_NAME" required:"yes"`
		Interval time.Duration `positional-arg-name:"INTERVAL" required:"yes"`
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

type DeviceOnuListImages struct {
	ListOutputOptions
	Args struct {
		Id DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceOnuDownloadImage struct {
	Args struct {
		Id           DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
		Name         string   `positional-arg-name:"IMAGE_NAME" required:"yes"`
		Url          string   `positional-arg-name:"IMAGE_URL" required:"yes"`
		ImageVersion string   `positional-arg-name:"IMAGE_VERSION" required:"yes"`
		Crc          uint32   `positional-arg-name:"IMAGE_CRC" required:"yes"`
		LocalDir     string   `positional-arg-name:"IMAGE_LOCAL_DIRECTORY"`
	} `positional-args:"yes"`
}

type DeviceOnuActivateImageUpdate struct {
	Args struct {
		Id           DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
		Name         string   `positional-arg-name:"IMAGE_NAME" required:"yes"`
		ImageVersion string   `positional-arg-name:"IMAGE_VERSION" required:"yes"`
		SaveConfig   bool     `positional-arg-name:"SAVE_EXISTING_CONFIG"`
		LocalDir     string   `positional-arg-name:"IMAGE_LOCAL_DIRECTORY"`
	} `positional-args:"yes"`
}

type OnuDownloadImage struct {
	ListOutputOptions
	Args struct {
		ImageVersion      string     `positional-arg-name:"IMAGE_VERSION" required:"yes"`
		Url               string     `positional-arg-name:"IMAGE_URL" required:"yes"`
		Vendor            string     `positional-arg-name:"IMAGE_VENDOR"`
		ActivateOnSuccess bool       `positional-arg-name:"IMAGE_ACTIVATE_ON_SUCCESS"`
		CommitOnSuccess   bool       `positional-arg-name:"IMAGE_COMMIT_ON_SUCCESS"`
		Crc               uint32     `positional-arg-name:"IMAGE_CRC"`
		IDs               []DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type OnuActivateImage struct {
	ListOutputOptions
	Args struct {
		ImageVersion    string     `positional-arg-name:"IMAGE_VERSION" required:"yes"`
		CommitOnSuccess bool       `positional-arg-name:"IMAGE_COMMIT_ON_SUCCESS"`
		IDs             []DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type OnuAbortUpgradeImage struct {
	ListOutputOptions
	Args struct {
		ImageVersion string     `positional-arg-name:"IMAGE_VERSION" required:"yes"`
		IDs          []DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type OnuCommitImage struct {
	ListOutputOptions
	Args struct {
		ImageVersion string     `positional-arg-name:"IMAGE_VERSION" required:"yes"`
		IDs          []DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type OnuImageStatus struct {
	ListOutputOptions
	Args struct {
		ImageVersion string     `positional-arg-name:"IMAGE_VERSION" required:"yes"`
		IDs          []DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type OnuListImages struct {
	ListOutputOptions
	Args struct {
		Id DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceGetPortStats struct {
	ListOutputOptions
	Args struct {
		Id       DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
		PortNo   uint32   `positional-arg-name:"PORT_NO" required:"yes"`
		PortType string   `positional-arg-name:"PORT_TYPE" required:"yes"`
	} `positional-args:"yes"`
}
type UniStatus struct {
	ListOutputOptions
	Args struct {
		Id       DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
		UniIndex uint32   `positional-arg-name:"UNI_INDEX" required:"yes"`
	} `positional-args:"yes"`
}
type OnuPonOpticalInfo struct {
	ListOutputOptions
	Args struct {
		Id DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type GetOnuStats struct {
	ListOutputOptions
	Args struct {
		OltId  DeviceId `positional-arg-name:"OLT_DEVICE_ID" required:"yes"`
		IntfId uint32   `positional-arg-name:"PON_INTF_ID" required:"yes"`
		OnuId  uint32   `positional-arg-name:"ONU_ID" required:"yes"`
	} `positional-args:"yes"`
}

type GetOnuEthernetFrameExtendedPmCounters struct {
	ListOutputOptions
	Reset bool `long:"reset" description:"Reset the counters"`
	Args  struct {
		Id       DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
		UniIndex *uint32  `positional-arg-name:"UNI_INDEX"`
	} `positional-args:"yes"`
}

type RxPower struct {
	ListOutputOptions
	Args struct {
		Id     DeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
		PortNo uint32   `positional-arg-name:"PORT_NO" required:"yes"`
		OnuNo  uint32   `positional-arg-name:"ONU_NO" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceOpts struct {
	List    DeviceList          `command:"list"`
	Create  DeviceCreate        `command:"create"`
	Delete  DeviceDelete        `command:"delete"`
	Enable  DeviceEnable        `command:"enable"`
	Disable DeviceDisable       `command:"disable"`
	Flows   DeviceFlowList      `command:"flows"`
	Groups  DeviceFlowGroupList `command:"groups"`
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
			List    DevicePmConfigGroupList         `command:"list"`
			Enable  DevicePmConfigGroupEnable       `command:"enable"`
			Disable DevicePmConfigGroupDisable      `command:"disable"`
			Set     DevicePmConfigGroupFrequencySet `command:"set"`
		} `command:"group"`
		GroupMetric struct {
			List DevicePmConfigGroupMetricList `command:"list"`
		} `command:"groupmetric"`
	} `command:"pmconfig"`
	Image struct {
		Get      DeviceOnuListImages          `command:"list"`
		Download DeviceOnuDownloadImage       `command:"download"`
		Activate DeviceOnuActivateImageUpdate `command:"activate"`
	} `command:"image"`
	DownloadImage struct {
		Download     OnuDownloadImage     `command:"download"`
		Activate     OnuActivateImage     `command:"activate"`
		Commit       OnuCommitImage       `command:"commit"`
		AbortUpgrade OnuAbortUpgradeImage `command:"abort"`
		Status       OnuImageStatus       `command:"status"`
		List         OnuListImages        `command:"list" `
	} `command:"onuimage"`
	GetExtVal struct {
		Stats                   DeviceGetPortStats                    `command:"portstats"`
		UniStatus               UniStatus                             `command:"unistatus"`
		OpticalInfo             OnuPonOpticalInfo                     `command:"onu_pon_optical_info"`
		OnuStats                GetOnuStats                           `command:"onu_stats"`
		EthernetFrameExtendedPm GetOnuEthernetFrameExtendedPmCounters `command:"ethernet_frame_extended_pm"`
		RxPower                 RxPower                               `command:"rxpower"`
	} `command:"getextval"`
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

func (options *DeviceFlowGroupList) Execute(args []string) error {
	grp := &GroupList{}
	grp.ListOutputOptions = options.ListOutputOptions
	grp.GroupListOptions = options.GroupListOptions
	grp.Args.Id = string(options.Args.Id)
	grp.Method = "device-groups"
	return grp.Execute(args)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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
			if _, have := groups[string(options.Args.Group)]; !have {
				return fmt.Errorf("Group Name '%s' does not exist", options.Args.Group)
			}
			if string(options.Args.Group) == group.GroupName && !group.Enabled {
				group.Enabled = true
				_, err := client.UpdateDevicePmConfigs(ctx, pmConfigs)
				if err != nil {
					return err
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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
			if _, have := groups[string(options.Args.Group)]; !have {
				return fmt.Errorf("Group Name '%s' does not exist", options.Args.Group)
			}

			if string(options.Args.Group) == group.GroupName && group.Enabled {
				group.Enabled = false
				_, err := client.UpdateDevicePmConfigs(ctx, pmConfigs)
				if err != nil {
					return err
				}
			}
		}
	} else {
		return fmt.Errorf("Device '%s' does not have Group Metrics", options.Args.Id)
	}
	return nil
}

func (options *DevicePmConfigGroupFrequencySet) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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
			if _, have := groups[string(options.Args.Group)]; !have {
				return fmt.Errorf("group name '%s' does not exist", options.Args.Group)
			}

			if string(options.Args.Group) == group.GroupName {
				if !group.Enabled {
					return fmt.Errorf("group '%s' is not enabled", options.Args.Group)
				}
				group.GroupFreq = uint32(options.Args.Interval.Seconds())
				_, err = client.UpdateDevicePmConfigs(ctx, pmConfigs)
				if err != nil {
					return err
				}
			}
		}
	} else {
		return fmt.Errorf("device '%s' does not have group metrics", options.Args.Id)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	pmConfigs, err := client.ListDevicePmConfigs(ctx, &id)
	if err != nil {
		return err
	}

	pmConfigs.DefaultFreq = uint32(options.Args.Interval.Seconds())

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

func (options *OnuDownloadImage) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	var devIDList []*common.ID
	for _, i := range options.Args.IDs {

		devIDList = append(devIDList, &common.ID{Id: string(i)})
	}

	downloadImage := voltha.DeviceImageDownloadRequest{
		DeviceId: devIDList,
		Image: &voltha.Image{
			Url:     options.Args.Url,
			Crc32:   options.Args.Crc,
			Vendor:  options.Args.Vendor,
			Version: options.Args.ImageVersion,
		},
		ActivateOnSuccess: options.Args.ActivateOnSuccess,
		CommitOnSuccess:   options.Args.CommitOnSuccess,
	}

	deviceImageResp, err := client.DownloadImageToDevice(ctx, &downloadImage)
	if err != nil {
		return err
	}

	outputFormat := GetCommandOptionWithDefault("onu-image-download", "format", ONU_IMAGE_STATUS_FORMAT)
	// Make sure json output prints an empty list, not "null"
	if deviceImageResp.DeviceImageStates == nil {
		deviceImageResp.DeviceImageStates = make([]*voltha.DeviceImageState, 0)
	}
	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      deviceImageResp.DeviceImageStates,
	}
	GenerateOutput(&result)
	return nil

}

func (options *OnuActivateImage) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	var devIDList []*common.ID
	for _, i := range options.Args.IDs {

		devIDList = append(devIDList, &common.ID{Id: string(i)})
	}

	downloadImage := voltha.DeviceImageRequest{
		DeviceId:        devIDList,
		Version:         options.Args.ImageVersion,
		CommitOnSuccess: options.Args.CommitOnSuccess,
	}

	deviceImageResp, err := client.ActivateImage(ctx, &downloadImage)
	if err != nil {
		return err
	}

	outputFormat := GetCommandOptionWithDefault("onu-image-activate", "format", ONU_IMAGE_STATUS_FORMAT)
	// Make sure json output prints an empty list, not "null"
	if deviceImageResp.DeviceImageStates == nil {
		deviceImageResp.DeviceImageStates = make([]*voltha.DeviceImageState, 0)
	}
	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      deviceImageResp.DeviceImageStates,
	}
	GenerateOutput(&result)

	return nil

}

func (options *OnuAbortUpgradeImage) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	var devIDList []*common.ID
	for _, i := range options.Args.IDs {

		devIDList = append(devIDList, &common.ID{Id: string(i)})
	}

	downloadImage := voltha.DeviceImageRequest{
		DeviceId: devIDList,
		Version:  options.Args.ImageVersion,
	}

	deviceImageResp, err := client.AbortImageUpgradeToDevice(ctx, &downloadImage)
	if err != nil {
		return err
	}

	outputFormat := GetCommandOptionWithDefault("onu-image-abort", "format", ONU_IMAGE_STATUS_FORMAT)
	// Make sure json output prints an empty list, not "null"
	if deviceImageResp.DeviceImageStates == nil {
		deviceImageResp.DeviceImageStates = make([]*voltha.DeviceImageState, 0)
	}
	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      deviceImageResp.DeviceImageStates,
	}
	GenerateOutput(&result)

	return nil

}

func (options *OnuCommitImage) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	var devIDList []*common.ID
	for _, i := range options.Args.IDs {

		devIDList = append(devIDList, &common.ID{Id: string(i)})
	}
	downloadImage := voltha.DeviceImageRequest{
		DeviceId: devIDList,
		Version:  options.Args.ImageVersion,
	}

	deviceImageResp, err := client.CommitImage(ctx, &downloadImage)
	if err != nil {
		return err
	}

	outputFormat := GetCommandOptionWithDefault("onu-image-commit", "format", ONU_IMAGE_STATUS_FORMAT)
	// Make sure json output prints an empty list, not "null"
	if deviceImageResp.DeviceImageStates == nil {
		deviceImageResp.DeviceImageStates = make([]*voltha.DeviceImageState, 0)
	}
	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      deviceImageResp.DeviceImageStates,
	}
	GenerateOutput(&result)

	return nil

}

func (options *OnuListImages) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	id := common.ID{Id: string(options.Args.Id)}

	onuImages, err := client.GetOnuImages(ctx, &id)
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("onu-image-list", "format", ONU_IMAGE_LIST_FORMAT)
	}

	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	//TODO orderby

	// Make sure json output prints an empty list, not "null"
	if onuImages.Items == nil {
		onuImages.Items = make([]*voltha.OnuImage, 0)
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      onuImages.Items,
	}

	GenerateOutput(&result)
	return nil

}

func (options *OnuImageStatus) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	var devIDList []*common.ID
	for _, i := range options.Args.IDs {

		devIDList = append(devIDList, &common.ID{Id: string(i)})
	}

	imageStatusReq := voltha.DeviceImageRequest{
		DeviceId: devIDList,
		Version:  options.Args.ImageVersion,
	}
	imageStatus, err := client.GetImageStatus(ctx, &imageStatusReq)
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-image-list", "format", ONU_IMAGE_STATUS_FORMAT)
	}

	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	//TODO orderby

	// Make sure json output prints an empty list, not "null"
	if imageStatus.DeviceImageStates == nil {
		imageStatus.DeviceImageStates = make([]*voltha.DeviceImageState, 0)
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      imageStatus.DeviceImageStates,
	}

	GenerateOutput(&result)
	return nil

}

func (options *DeviceOnuListImages) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	id := common.ID{Id: string(options.Args.Id)}

	imageDownloads, err := client.ListImageDownloads(ctx, &id)
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-image-list", "format", DEFAULT_DEVICE_IMAGE_LIST_GET_FORMAT)
	}

	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	//TODO orderby

	// Make sure json output prints an empty list, not "null"
	if imageDownloads.Items == nil {
		imageDownloads.Items = make([]*voltha.ImageDownload, 0)
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      imageDownloads.Items,
	}

	GenerateOutput(&result)
	return nil

}

func (options *DeviceOnuDownloadImage) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	downloadImage := voltha.ImageDownload{
		Id:       string(options.Args.Id),
		Name:     options.Args.Name,
		Url:      options.Args.Url,
		Crc:      options.Args.Crc,
		LocalDir: options.Args.LocalDir,
	}

	_, err = client.DownloadImage(ctx, &downloadImage)
	if err != nil {
		return err
	}

	return nil

}

func (options *DeviceOnuActivateImageUpdate) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	downloadImage := voltha.ImageDownload{
		Id:           string(options.Args.Id),
		Name:         options.Args.Name,
		ImageVersion: options.Args.ImageVersion,
		SaveConfig:   options.Args.SaveConfig,
		LocalDir:     options.Args.LocalDir,
	}

	_, err = client.ActivateImageUpdate(ctx, &downloadImage)
	if err != nil {
		return err
	}

	return nil

}

type ReturnValueRow struct {
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
}

func (options *DeviceGetPortStats) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := extension.NewExtensionClient(conn)
	var portType extension.GetOltPortCounters_PortType

	if options.Args.PortType == "pon" {
		portType = extension.GetOltPortCounters_Port_PON_OLT
	} else if options.Args.PortType == "nni" {

		portType = extension.GetOltPortCounters_Port_ETHERNET_NNI
	} else {
		return fmt.Errorf("expected interface type pon/nni, provided %s", options.Args.PortType)
	}

	singleGetValReq := extension.SingleGetValueRequest{
		TargetId: string(options.Args.Id),
		Request: &extension.GetValueRequest{
			Request: &extension.GetValueRequest_OltPortInfo{
				OltPortInfo: &extension.GetOltPortCounters{
					PortNo:   options.Args.PortNo,
					PortType: portType,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()
	rv, err := client.GetExtValue(ctx, &singleGetValReq)
	if err != nil {
		Error.Printf("Error getting value on device Id %s,err=%s\n", options.Args.Id, ErrorToString(err))
		return err
	}

	if rv.Response.Status != extension.GetValueResponse_OK {
		return fmt.Errorf("failed to get port stats %v", rv.Response.ErrReason.String())
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-get-port-status", "format", DEFAULT_DEVICE_GET_PORT_STATUS_FORMAT)
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      rv.GetResponse().GetPortCoutners(),
	}
	GenerateOutput(&result)
	return nil
}

func (options *GetOnuStats) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := extension.NewExtensionClient(conn)

	singleGetValReq := extension.SingleGetValueRequest{
		TargetId: string(options.Args.OltId),
		Request: &extension.GetValueRequest{
			Request: &extension.GetValueRequest_OnuPonInfo{
				OnuPonInfo: &extension.GetOnuCountersRequest{
					IntfId: options.Args.IntfId,
					OnuId:  options.Args.OnuId,
				},
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()
	rv, err := client.GetExtValue(ctx, &singleGetValReq)
	if err != nil {
		Error.Printf("Error getting value on device Id %s,err=%s\n", options.Args.OltId, ErrorToString(err))
		return err
	}

	if rv.Response.Status != extension.GetValueResponse_OK {
		return fmt.Errorf("failed to get onu stats %v", rv.Response.ErrReason.String())
	}
	outputFormat := CharReplacer.Replace(options.Format)
	data, formatStr := buildOnuStatsOutputFormat(rv.GetResponse().GetOnuPonCounters())
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-get-onu-status", "format", formatStr)
	}
	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}
	GenerateOutput(&result)
	return nil
}

func (options *GetOnuEthernetFrameExtendedPmCounters) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := extension.NewExtensionClient(conn)
	var singleGetValReq extension.SingleGetValueRequest

	if options.Args.UniIndex != nil {
		singleGetValReq = extension.SingleGetValueRequest{
			TargetId: string(options.Args.Id),
			Request: &extension.GetValueRequest{
				Request: &extension.GetValueRequest_OnuInfo{
					OnuInfo: &extension.GetOmciEthernetFrameExtendedPmRequest{
						OnuDeviceId: string(options.Args.Id),
						Reset_:      options.Reset,
						IsUniIndex: &extension.GetOmciEthernetFrameExtendedPmRequest_UniIndex{
							UniIndex: *options.Args.UniIndex,
						},
					},
				},
			},
		}
	} else {
		singleGetValReq = extension.SingleGetValueRequest{
			TargetId: string(options.Args.Id),
			Request: &extension.GetValueRequest{
				Request: &extension.GetValueRequest_OnuInfo{
					OnuInfo: &extension.GetOmciEthernetFrameExtendedPmRequest{
						OnuDeviceId: string(options.Args.Id),
						Reset_:      options.Reset,
					},
				},
			},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()
	rv, err := client.GetExtValue(ctx, &singleGetValReq)
	if err != nil {
		Error.Printf("Error getting value on device Id %s,err=%s\n", options.Args.Id, ErrorToString(err))
		return err
	}

	if rv.Response.Status != extension.GetValueResponse_OK {
		return fmt.Errorf("failed to get ethernet frame extended pm counters %v", rv.Response.ErrReason.String())
	}
	outputFormat := CharReplacer.Replace(options.Format)
	data := buildOnuEthernetFrameExtendedPmOutputFormat(rv.GetResponse().GetOnuCounters())
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-get-onu-status", "format", DEFAULT_ETHERNET_FRAME_EXTENDED_PM_COUNTERS_FORMAT)
	}
	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}
	GenerateOutput(&result)
	return nil
}

func (options *UniStatus) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := extension.NewExtensionClient(conn)

	singleGetValReq := extension.SingleGetValueRequest{
		TargetId: string(options.Args.Id),
		Request: &extension.GetValueRequest{
			Request: &extension.GetValueRequest_UniInfo{
				UniInfo: &extension.GetOnuUniInfoRequest{
					UniIndex: options.Args.UniIndex,
				},
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()
	rv, err := client.GetExtValue(ctx, &singleGetValReq)
	if err != nil {
		Error.Printf("Error getting value on device Id %s,err=%s\n", options.Args.Id, ErrorToString(err))
		return err
	}
	if rv.Response.Status != extension.GetValueResponse_OK {
		return fmt.Errorf("failed to get uni status %v", rv.Response.ErrReason.String())
	}
	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-get-uni-status", "format", DEFAULT_DEVICE_GET_UNI_STATUS_FORMAT)
	}
	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      rv.GetResponse().GetUniInfo(),
	}
	GenerateOutput(&result)
	return nil
}

func (options *OnuPonOpticalInfo) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := extension.NewExtensionClient(conn)

	singleGetValReq := extension.SingleGetValueRequest{
		TargetId: string(options.Args.Id),
		Request: &extension.GetValueRequest{
			Request: &extension.GetValueRequest_OnuOpticalInfo{
				OnuOpticalInfo: &extension.GetOnuPonOpticalInfo{},
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()
	rv, err := client.GetExtValue(ctx, &singleGetValReq)
	if err != nil {
		Error.Printf("Error getting value on device Id %s,err=%s\n", options.Args.Id, ErrorToString(err))
		return err
	}
	if rv.Response.Status != extension.GetValueResponse_OK {
		return fmt.Errorf("failed to get onu pon optical info %v", rv.Response.ErrReason.String())
	}
	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-get-onu-pon-optical-info", "format", DEFAULT_ONU_PON_OPTICAL_INFO_STATUS_FORMAT)
	}
	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      rv.GetResponse().GetOnuOpticalInfo(),
	}
	GenerateOutput(&result)
	return nil
}

func (options *RxPower) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := extension.NewExtensionClient(conn)

	singleGetValReq := extension.SingleGetValueRequest{
		TargetId: string(options.Args.Id),
		Request: &extension.GetValueRequest{
			Request: &extension.GetValueRequest_RxPower{
				RxPower: &extension.GetRxPowerRequest{
					IntfId: options.Args.PortNo,
					OnuId:  options.Args.OnuNo,
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()
	rv, err := client.GetExtValue(ctx, &singleGetValReq)
	if err != nil {
		Error.Printf("Error getting value on device Id %s,err=%s\n", options.Args.Id, ErrorToString(err))
		return err
	}
	if rv.Response.Status != extension.GetValueResponse_OK {
		return fmt.Errorf("failed to get rx power %v", rv.Response.ErrReason.String())
	}
	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-get-rx-power", "format", DEFAULT_RX_POWER_STATUS_FORMAT)
	}
	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      rv.GetResponse().GetRxPower(),
	}
	GenerateOutput(&result)
	return nil
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

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
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
