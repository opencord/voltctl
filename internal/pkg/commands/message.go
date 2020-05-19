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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/pkg/filter"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltha-protos/v3/go/inter_container"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

/*
 * The "message listen" command supports two types of output:
 *    1) A summary output where a row is displayed for each message received. For the summary
 *       format, DEFAULT_MESSAGE_FORMAT contains the default list of columns that will be
 *       display and can be overridden at runtime.
 *    2) A body output where the full grpcurl or json body is output for each message received.
 *
 * These two modes are switched by using the "-b" / "--body" flag.
 *
 * The summary mode has the potential to aggregate data together from multiple parts of the
 * message. For example, it currently aggregates the InterAdapterHeader contents together with
 * the InterContainerHeader contents.
 *
 * Similar to "event listen", the  "message listen" command operates in a streaming mode, rather
 * than collecting a list of results and then emitting them at program exit. This is done to
 * facilitate options such as "-F" / "--follow" where the program is intended to  operate
 * continuously. This means that automatically calculating column widths is not practical, and
 * a set of Fixed widths (MessageHeaderDefaultWidths) are predefined.
 *
 * As there are multiple kafka topics that can be listened to, specifying a topic is a
 * mandatory positional argument for the `message listen` command. Common topics include:
 *   * openolt
 *   * brcm_openonu_adapter
 *   * rwcore
 *   * core-pair-1
 */

const (
	DEFAULT_MESSAGE_FORMAT = "table{{.Id}}\t{{.Type}}\t{{.FromTopic}}\t{{.ToTopic}}\t{{.KeyTopic}}\t{{.InterAdapterType}}"
)

type MessageListenOpts struct {
	Format string `long:"format" value-name:"FORMAT" default:"" description:"Format to use to output structured data"`
	// nolint: staticcheck
	OutputAs string `short:"o" long:"outputas" default:"table" choice:"table" choice:"json" choice:"yaml" description:"Type of output to generate"`
	Filter   string `short:"f" long:"filter" default:"" value-name:"FILTER" description:"Only display results that match filter"`
	Follow   bool   `short:"F" long:"follow" description:"Continue to consume until CTRL-C is pressed"`
	ShowBody bool   `short:"b" long:"show-body" description:"Show body of messages rather than only a header summary"`
	Count    int    `short:"c" long:"count" default:"-1" value-name:"LIMIT" description:"Limit the count of messages that will be printed"`
	Now      bool   `short:"n" long:"now" description:"Stop printing messages when current time is reached"`
	Timeout  int    `short:"t" long:"idle" default:"900" value-name:"SECONDS" description:"Timeout if no message received within specified seconds"`
	Since    string `short:"s" long:"since" default:"" value-name:"TIMESTAMP" description:"Do not show entries before timestamp"`

	Args struct {
		Topic string
	} `positional-args:"yes" required:"yes"`
}

type MessageOpts struct {
	MessageListen MessageListenOpts `command:"listen"`
}

var interAdapterOpts = MessageOpts{}

/* MessageHeader is a set of fields extracted
 * from voltha.MessageHeader as well as useful other
 * places such as InterAdapterHeader. These are fields that
 * will be summarized in list mode and/or can be filtered
 * on.
 */
type MessageHeader struct {
	Id               string    `json:"id"`
	Type             string    `json:"type"`
	FromTopic        string    `json:"from_topic"`
	ToTopic          string    `json:"to_topic"`
	KeyTopic         string    `json:"key_topic"`
	Timestamp        time.Time `json:"timestamp"`
	InterAdapterType string    `json:"inter_adapter_type"` // interadapter header
	ToDeviceId       string    `json:"to_device_id"`       // interadapter header
	ProxyDeviceId    string    `json:"proxy_device_id"`    //interadapter header
}

/* Fixed widths because we output in a continuous streaming
 * mode rather than a table-based dump at the end.
 */
type MessageHeaderWidths struct {
	Id               int
	Type             int
	FromTopic        int
	ToTopic          int
	KeyTopic         int
	InterAdapterType int
	ToDeviceId       int
	ProxyDeviceId    int
	Timestamp        int
}

var DefaultMessageWidths MessageHeaderWidths = MessageHeaderWidths{
	Id:               32,
	Type:             10,
	FromTopic:        16,
	ToTopic:          16,
	KeyTopic:         10,
	Timestamp:        10,
	InterAdapterType: 14,
	ToDeviceId:       10,
	ProxyDeviceId:    10,
}

// jsonpb requires a resolver to resolve Any.Any into proto.Message.
type VolthaAnyResolver struct{}

func (r *VolthaAnyResolver) Resolve(typeURL string) (proto.Message, error) {
	// TODO: We should be able to get this automatically via reflection using
	// the following commented-out code, but it requires upgrading voltctl to
	// use newer versions of protobuf libraries.

	/*
		msgType, err := protoregistry.GlobalTypes.FindMessageByURL(typeURL)
		if err != nil {
			return err
		}
		return msgType.New(), nil*/

	// The intercontianer message bus is where we need to map from Any.Any
	// to the appropriate protos when generating json output.

	typeURL = strings.TrimPrefix(typeURL, "type.googleapis.com/")

	switch typeURL {
	case "voltha.StrType":
		return &inter_container.StrType{}, nil
	case "voltha.IntType":
		return &inter_container.IntType{}, nil
	case "voltha.BoolType":
		return &inter_container.BoolType{}, nil
	case "voltha.Packet":
		return &inter_container.Packet{}, nil
	case "voltha.ErrorCode":
		return &inter_container.ErrorCode{}, nil
	case "voltha.Error":
		return &inter_container.Error{}, nil
	case "voltha.Header":
		return &inter_container.Header{}, nil
	case "voltha.Argument":
		return &inter_container.Argument{}, nil
	case "voltha.InterContainerMessage":
		return &inter_container.InterContainerMessage{}, nil
	case "voltha.InterContainerRequestBody":
		return &inter_container.InterContainerRequestBody{}, nil
	case "voltha.InterContainerResponseBody":
		return &inter_container.InterContainerResponseBody{}, nil
	case "voltha.SwitchCapability":
		return &inter_container.SwitchCapability{}, nil
	case "voltha.PortCapability":
		return &inter_container.PortCapability{}, nil
	case "voltha.DeviceDiscovered":
		return &inter_container.DeviceDiscovered{}, nil
	case "voltha.InterAdapterMessageType":
		return &inter_container.InterAdapterMessageType{}, nil
	case "voltha.InterAdapterOmciMessage":
		return &inter_container.InterAdapterOmciMessage{}, nil
	case "voltha.InterAdapterTechProfileDownloadMessage":
		return &inter_container.InterAdapterTechProfileDownloadMessage{}, nil
	case "voltha.InterAdapterDeleteGemPortMessage":
		return &inter_container.InterAdapterDeleteGemPortMessage{}, nil
	case "voltha.InterAdapterDeleteTcontMessage":
		return &inter_container.InterAdapterDeleteTcontMessage{}, nil
	case "voltha.InterAdapterResponseBody":
		return &inter_container.InterAdapterResponseBody{}, nil
	case "voltha.InterAdapterMessage":
		return &inter_container.InterAdapterMessage{}, nil
	}

	return nil, fmt.Errorf("Unknown any type: %s", typeURL)
}

func RegisterMessageCommands(parent *flags.Parser) {
	if _, err := parent.AddCommand("message", "message commands", "Commands for observing messages between components", &interAdapterOpts); err != nil {
		Error.Fatalf("Unable to register message commands with voltctl command parser: %s", err.Error())
	}
}

// Extract the header, as well as a few other items that might be of interest
func DecodeInterContainerHeader(b []byte, ts time.Time) (*MessageHeader, error) {
	m := &inter_container.InterContainerMessage{}
	if err := proto.Unmarshal(b, m); err != nil {
		return nil, err
	}

	header := m.Header
	id := header.Id
	msgType := header.Type
	fromTopic := header.FromTopic
	toTopic := header.ToTopic
	keyTopic := header.KeyTopic
	timestamp, err := DecodeTimestamp(header.Timestamp)
	if err != nil {
		return nil, err
	}

	// Pull some additional data out of the InterAdapterHeader, if it
	// is embedded inside the InterContainerMessage

	var iaMessageTypeStr string
	var toDeviceId string
	var proxyDeviceId string

	bodyKind, err := ptypes.AnyMessageName(m.Body)
	if err != nil {
		return nil, err
	}

	switch bodyKind {
	case "voltha.InterContainerRequestBody":
		icRequest := &inter_container.InterContainerRequestBody{}
		err := ptypes.UnmarshalAny(m.Body, icRequest)
		if err != nil {
			return nil, err
		}

		argList := icRequest.Args
		for _, arg := range argList {
			key := arg.Key
			if key == "msg" {
				argBodyKind, err := ptypes.AnyMessageName(m.Body)
				if err != nil {
					return nil, err
				}
				switch argBodyKind {
				case "voltha.InterAdapterMessage":
					iaMsg := &inter_container.InterAdapterMessage{}
					err := ptypes.UnmarshalAny(arg.Value, iaMsg)
					if err != nil {
						return nil, err
					}
					iaHeader := iaMsg.Header
					iaMessageType := iaHeader.Type
					iaMessageTypeStr = inter_container.InterAdapterMessageType_Types_name[int32(iaMessageType)]

					toDeviceId = iaHeader.ToDeviceId
					proxyDeviceId = iaHeader.ProxyDeviceId
				}
			}
		}
	}

	messageHeaderType := inter_container.MessageType_name[int32(msgType)]

	icHeader := MessageHeader{Id: id,
		Type:             messageHeaderType,
		FromTopic:        fromTopic,
		ToTopic:          toTopic,
		KeyTopic:         keyTopic,
		Timestamp:        timestamp,
		InterAdapterType: iaMessageTypeStr,
		ProxyDeviceId:    proxyDeviceId,
		ToDeviceId:       toDeviceId,
	}

	return &icHeader, nil
}

// Print the full message, either in JSON or in GRPCURL-human-readable format,
// depending on which grpcurl formatter is passed in.
func PrintInterContainerMessage(outputAs string, b []byte) error {
	ms := &inter_container.InterContainerMessage{}
	if err := proto.Unmarshal(b, ms); err != nil {
		return err
	}

	if outputAs == "json" {
		marshaler := jsonpb.Marshaler{EmitDefaults: true, AnyResolver: &VolthaAnyResolver{}}
		asJson, err := marshaler.MarshalToString(ms)
		if err != nil {
			return fmt.Errorf("Failed to marshal the json: %s", err)
		}
		fmt.Println(asJson)
	} else {
		// print in golang native format
		fmt.Printf("%v\n", ms)
	}

	return nil
}

// Print just the enriched InterContainerHeader. This is either in JSON format, or in
// table format.
func PrintInterContainerHeader(outputAs string, outputFormat string, hdr *MessageHeader) error {
	if outputAs == "json" {
		asJson, err := json.Marshal(hdr)
		if err != nil {
			return fmt.Errorf("Error marshalling JSON: %v", err)
		} else {
			fmt.Printf("%s\n", asJson)
		}
	} else {
		f := format.Format(outputFormat)
		output, err := f.ExecuteFixedWidth(DefaultMessageWidths, false, *hdr)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", output)
	}
	return nil
}

// Start output, print any column headers or other start characters
func (options *MessageListenOpts) StartOutput(outputFormat string) error {
	if options.OutputAs == "json" {
		fmt.Println("[")
	} else if (options.OutputAs == "table") && !options.ShowBody {
		f := format.Format(outputFormat)
		output, err := f.ExecuteFixedWidth(DefaultMessageWidths, true, nil)
		if err != nil {
			return err
		}
		fmt.Println(output)
	}
	return nil
}

// Finish output, print any column footers or other end characters
func (options *MessageListenOpts) FinishOutput() {
	if options.OutputAs == "json" {
		fmt.Println("]")
	}
}

func (options *MessageListenOpts) Execute(args []string) error {
	ProcessGlobalOptions()
	if GlobalConfig.Kafka == "" {
		return errors.New("Kafka address is not specified")
	}

	config := sarama.NewConfig()
	config.ClientID = "go-kafka-consumer"
	config.Consumer.Return.Errors = true
	config.Version = sarama.V1_0_0_0
	brokers := []string{GlobalConfig.Kafka}

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return err
	}

	defer func() {
		if err := client.Close(); err != nil {
			panic(err)
		}
	}()

	consumer, consumerErrors, highwaterMarks, err := startInterContainerConsumer([]string{options.Args.Topic}, client)
	if err != nil {
		return err
	}

	highwater := highwaterMarks[options.Args.Topic]

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Count how many message processed
	consumeCount := 0

	// Count how many messages were printed
	count := 0

	var headerFilter *filter.Filter
	if options.Filter != "" {
		headerFilterVal, err := filter.Parse(options.Filter)
		if err != nil {
			return fmt.Errorf("Failed to parse filter: %v", err)
		}
		headerFilter = &headerFilterVal
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("intercontainer-listen", "format", DEFAULT_MESSAGE_FORMAT)
	}

	err = options.StartOutput(outputFormat)
	if err != nil {
		return err
	}

	var since *time.Time
	if options.Since != "" {
		since, err = ParseSince(options.Since)
		if err != nil {
			return err
		}
	}

	// Get signnal for finish
	doneCh := make(chan struct{})
	go func() {
		tStart := time.Now()
	Loop:
		for {
			// Initialize the idle timeout timer
			timeoutTimer := time.NewTimer(time.Duration(options.Timeout) * time.Second)
			select {
			case msg := <-consumer:
				consumeCount++
				hdr, err := DecodeInterContainerHeader(msg.Value, msg.Timestamp)
				if err != nil {
					log.Printf("Error decoding header %v\n", err)
					continue
				}
				if headerFilter != nil && !headerFilter.Evaluate(*hdr) {
					// skip printing message
				} else if since != nil && hdr.Timestamp.Before(*since) {
					// it's too old
				} else {
					// comma separated between this message and predecessor
					if count > 0 {
						if options.OutputAs == "json" {
							fmt.Println(",")
						}
					}

					// Print it
					if options.ShowBody {
						if err := PrintInterContainerMessage(options.OutputAs, msg.Value); err != nil {
							log.Printf("%v\n", err)
						}
					} else {
						if err := PrintInterContainerHeader(options.OutputAs, outputFormat, hdr); err != nil {
							log.Printf("%v\n", err)
						}
					}

					// Check to see if we've hit the "count" threshold the user specified
					count++
					if (options.Count > 0) && (count >= options.Count) {
						log.Println("Count reached")
						doneCh <- struct{}{}
						break Loop
					}

					// Check to see if we've hit the "now" threshold the user specified
					if (options.Now) && (!hdr.Timestamp.Before(tStart)) {
						log.Println("Now timestamp reached")
						doneCh <- struct{}{}
						break Loop
					}
				}

				// If we're not in follow mode, see if we hit the highwater mark
				if !options.Follow && !options.Now && (msg.Offset >= highwater) {
					log.Println("High water reached")
					doneCh <- struct{}{}
					break Loop
				}

				// Reset the timeout timer
				if !timeoutTimer.Stop() {
					<-timeoutTimer.C
				}
			case consumerError := <-consumerErrors:
				log.Printf("Received consumerError topic=%v, partition=%v, err=%v\n", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				doneCh <- struct{}{}
			case <-signals:
				doneCh <- struct{}{}
			case <-timeoutTimer.C:
				log.Printf("Idle timeout\n")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh

	options.FinishOutput()

	log.Printf("Consumed %d messages. Printed %d messages", consumeCount, count)

	return nil
}

// Consume message from Sarama and send them out on a channel.
// Supports multiple topics.
// Taken from Sarama example consumer.
func startInterContainerConsumer(topics []string, client sarama.Client) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError, map[string]int64, error) {
	master, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return nil, nil, nil, err
	}

	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)
	highwater := make(map[string]int64)
	for _, topic := range topics {
		if strings.Contains(topic, "__consumer_offsets") {
			continue
		}
		partitions, _ := master.Partitions(topic)

		// TODO: Add support for multiple partitions
		if len(partitions) > 1 {
			log.Printf("WARNING: %d partitions on topic %s but we only listen to the first one\n", len(partitions), topic)
		}

		hw, err := client.GetOffset("openolt", partitions[0], sarama.OffsetNewest)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error in consume() getting highwater: Topic %v Partitions: %v", topic, partitions)
		}
		highwater[topic] = hw

		consumer, err := master.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
		if nil != err {
			return nil, nil, nil, fmt.Errorf("Error in consume(): Topic %v Partitions: %v", topic, partitions)
		}
		log.Println(" Start consuming topic ", topic)
		go func(topic string, consumer sarama.PartitionConsumer) {
			for {
				select {
				case consumerError := <-consumer.Errors():
					errors <- consumerError

				case msg := <-consumer.Messages():
					consumers <- msg
				}
			}
		}(topic, consumer)
	}

	return consumers, errors, highwater, nil
}
