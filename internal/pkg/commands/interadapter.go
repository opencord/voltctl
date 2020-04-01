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
	"github.com/fullstorydev/grpcurl"
	//	"github.com/golang/protobuf/ptypes"
	//	"github.com/golang/protobuf/ptypes/timestamp"
	//"github.com/golang/protobuf/ptypes/any"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/opencord/voltctl/pkg/filter"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltctl/pkg/model"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	DEFAULT_INTER_CONTAINER_FORMAT = "table{{.Id}}\t{{.Type}}\t{{.FromTopic}}\t{{.ToTopic}}\t{{.KeyTopic}}"
)

type InterContainerListenOpts struct {
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

type InterContainerOpts struct {
	InterContainerListen InterContainerListenOpts `command:"listen"`
}

var interAdapterOpts = InterContainerOpts{}

type InterContainerHeader struct {
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

type InterContainerHeaderWidths struct {
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

var DefaultInterContainerWidths InterContainerHeaderWidths = InterContainerHeaderWidths{
	Id:               32,
	Type:             10,
	FromTopic:        16,
	ToTopic:          16,
	KeyTopic:         10,
	Timestamp:        10,
	InterAdapterType: 10,
	ToDeviceId:       10,
	ProxyDeviceId:    10,
}

func RegisterInterContainerCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("intercontainer", "intercontainer commands", "Commands for observing intercontainer messages", &interAdapterOpts)
	if err != nil {
		Error.Fatalf("Unable to register intercontainer commands with voltctl command parser: %s", err.Error())
	}
}

// Extract the header, as well as a few other items that might be of interest
func DecodeInterContainerHeader(md *desc.MessageDescriptor, b []byte, ts time.Time, f grpcurl.Formatter) (*InterContainerHeader, error) {
	m := dynamic.NewMessage(md)
	err := m.Unmarshal(b)
	if err != nil {
		return nil, err
	}

	headerIntf, err := m.TryGetFieldByName("header")
	if err != nil {
		return nil, err
	}

	header := headerIntf.(*dynamic.Message)

	idIntf, err := header.TryGetFieldByName("id")
	if err != nil {
		return nil, err
	}
	id := idIntf.(string)

	typeIntf, err := header.TryGetFieldByName("type")
	if err != nil {
		return nil, err
	}
	msgType := typeIntf.(int32)

	fromTopicIntf, err := header.TryGetFieldByName("from_topic")
	if err != nil {
		return nil, err
	}
	fromTopic := fromTopicIntf.(string)

	toTopicIntf, err := header.TryGetFieldByName("to_topic")
	if err != nil {
		return nil, err
	}
	toTopic := toTopicIntf.(string)

	keyTopicIntf, err := header.TryGetFieldByName("key_topic")
	if err != nil {
		return nil, err
	}
	keyTopic := keyTopicIntf.(string)

	/*
		proxyDeviceIdIntf, err := header.TryGetFieldByName("proxy_device_id")
		if err != nil {
			return nil, err
		}
		proxyDeviceId := proxyDeviceIdIntf.(string)
	*/

	timestampIntf, err := header.TryGetFieldByName("timestamp")
	if err != nil {
		return nil, err
	}
	timestamp, err := DecodeTimestamp(timestampIntf)
	if err != nil {
		return nil, err
	}

	iaHeader := InterContainerHeader{Id: id,
		Type:      model.GetEnumString(header, "type", msgType),
		FromTopic: fromTopic,
		ToTopic:   toTopic,
		KeyTopic:  keyTopic,
		Timestamp: timestamp}

	return &iaHeader, nil
}

// Print the full message, either in JSON or in GRPCURL-human-readable format,
// depending on which grpcurl formatter is passed in.
func PrintInterContainerMessage(f grpcurl.Formatter, md *desc.MessageDescriptor, b []byte) error {
	m := dynamic.NewMessage(md)
	err := m.Unmarshal(b)
	if err != nil {
		return err
	}
	s, err := f(m)
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

// Print just the enriched InterContainerHeader. This is either in JSON format, or in
// table format.
func PrintInterContainerHeader(outputAs string, outputFormat string, hdr *InterContainerHeader) error {
	if outputAs == "json" {
		asJson, err := json.Marshal(hdr)
		if err != nil {
			return fmt.Errorf("Error marshalling JSON: %v", err)
		} else {
			fmt.Printf("%s\n", asJson)
		}
	} else {
		f := format.Format(outputFormat)
		output, err := f.ExecuteFixedWidth(DefaultInterContainerWidths, false, *hdr)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", output)
	}
	return nil
}

func GetInterContainerMessageDesc() (*desc.MessageDescriptor, error) {
	// This is a very long-winded way to get a message descriptor

	descriptor, err := GetDescriptorSource()
	if err != nil {
		return nil, err
	}

	// get the symbol for voltha.InterContainerMessage
	iaSymbol, err := descriptor.FindSymbol("voltha.InterContainerMessage")
	if err != nil {
		return nil, err
	}

	/*
	 * iaSymbol is a Descriptor, but not a MessageDescrptior,
	 * so we can't look at it's fields yet. Go back to the file,
	 * call FindMessage to get the Message, then ...
	 */

	iaFile := iaSymbol.GetFile()
	iaMessage := iaFile.FindMessage("voltha.InterContainerMessage")

	return iaMessage, nil
}

// Start output, print any column headers or other start characters
func (options *InterContainerListenOpts) StartOutput(outputFormat string) error {
	if options.OutputAs == "json" {
		fmt.Println("[")
	} else if (options.OutputAs == "table") && !options.ShowBody {
		f := format.Format(outputFormat)
		output, err := f.ExecuteFixedWidth(DefaultInterContainerWidths, true, nil)
		if err != nil {
			return err
		}
		fmt.Println(output)
	}
	return nil
}

// Finish output, print any column footers or other end characters
func (options *InterContainerListenOpts) FinishOutput() {
	if options.OutputAs == "json" {
		fmt.Println("]")
	}
}

func (options *InterContainerListenOpts) Execute(args []string) error {
	ProcessGlobalOptions()
	if GlobalConfig.Kafka == "" {
		return errors.New("Kafka address is not specified")
	}

	iaMessage, err := GetInterContainerMessageDesc()
	if err != nil {
		return err
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

	var grpcurlFormatter grpcurl.Formatter
	// need a descriptor source, any method will do
	descriptor, _, err := GetMethod("device-list")
	if err != nil {
		return err
	}

	jsonFormatter := grpcurl.NewJSONFormatter(false, grpcurl.AnyResolverFromDescriptorSource(descriptor))
	if options.ShowBody {
		if options.OutputAs == "json" {
			grpcurlFormatter = jsonFormatter
		} else {
			grpcurlFormatter = grpcurl.NewTextFormatter(false)
		}
	}

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
		outputFormat = GetCommandOptionWithDefault("intercontainer-listen", "format", DEFAULT_INTER_CONTAINER_FORMAT)
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
				hdr, err := DecodeInterContainerHeader(iaMessage, msg.Value, msg.Timestamp, jsonFormatter)
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
						if err := PrintInterContainerMessage(grpcurlFormatter, iaMessage, msg.Value); err != nil {
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
