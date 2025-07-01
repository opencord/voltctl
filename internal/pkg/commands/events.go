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
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck
	"github.com/golang/protobuf/proto"  //nolint:staticcheck
	"github.com/golang/protobuf/ptypes/timestamp"
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/pkg/filter"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltha-protos/v5/go/voltha"
)

const (
	DEFAULT_EVENT_FORMAT = "table{{.Category}}\t{{.SubCategory}}\t{{.Type}}\t{{.Timestamp}}\t{{.Device_ids}}\t{{.Titles}}"
)

type EventListenOpts struct {
	Format string `long:"format" value-name:"FORMAT" default:"" description:"Format to use to output structured data"`
	// nolint: staticcheck
	OutputAs string `short:"o" long:"outputas" default:"table" choice:"table" choice:"json" choice:"yaml" description:"Type of output to generate"`
	Filter   string `short:"f" long:"filter" default:"" value-name:"FILTER" description:"Only display results that match filter"`
	Follow   bool   `short:"F" long:"follow" description:"Continue to consume until CTRL-C is pressed"`
	ShowBody bool   `short:"b" long:"show-body" description:"Show body of events rather than only a header summary"`
	Count    int    `short:"c" long:"count" default:"-1" value-name:"LIMIT" description:"Limit the count of messages that will be printed"`
	Now      bool   `short:"n" long:"now" description:"Stop printing events when current time is reached"`
	Timeout  int    `short:"t" long:"idle" default:"900" value-name:"SECONDS" description:"Timeout if no message received within specified seconds"`
	Since    string `short:"s" long:"since" default:"" value-name:"TIMESTAMP" description:"Do not show entries before timestamp"`
}

type EventOpts struct {
	EventListen EventListenOpts `command:"listen"`
}

var eventOpts = EventOpts{}

type EventHeader struct {
	Category    string    `json:"category"`
	SubCategory string    `json:"sub_category"`
	Type        string    `json:"type"`
	Raised_ts   time.Time `json:"raised_ts"`
	Reported_ts time.Time `json:"reported_ts"`
	Device_ids  []string  `json:"device_ids"` // Opportunistically collected list of device_ids
	Titles      []string  `json:"titles"`     // Opportunistically collected list of titles
	Timestamp   time.Time `json:"timestamp"`  // Timestamp from Kafka
}

type EventHeaderWidths struct {
	Category    int
	SubCategory int
	Type        int
	Raised_ts   int
	Reported_ts int
	Device_ids  int
	Titles      int
	Timestamp   int
}

var DefaultWidths EventHeaderWidths = EventHeaderWidths{
	Category:    13,
	SubCategory: 3,
	Type:        12,
	Raised_ts:   10,
	Reported_ts: 10,
	Device_ids:  40,
	Titles:      40,
	Timestamp:   10,
}

func RegisterEventCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("event", "event commands", "Commands for observing events", &eventOpts)
	if err != nil {
		Error.Fatalf("Unable to register event commands with voltctl command parser: %s", err.Error())
	}
}

func ParseSince(s string) (*time.Time, error) {
	if strings.EqualFold(s, "now") {
		since := time.Now()
		return &since, nil
	}

	rfc3339Time, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return &rfc3339Time, nil
	}

	duration, err := time.ParseDuration(s)
	if err == nil {
		since := time.Now().Add(-duration)
		return &since, nil
	}

	return nil, fmt.Errorf("Unable to parse time specification `%s`. Please use either `now`, a duration, or an RFC3339-compliant string", s)
}

// Convert a timestamp field in an event to a time.Time
func DecodeTimestamp(tsIntf interface{}) (time.Time, error) {
	ts, okay := tsIntf.(*timestamp.Timestamp)
	if okay {
		// Voltha-Protos 3.2.3 and above
		return ts.AsTime(), nil
	}
	tsFloat, okay := tsIntf.(float32)
	if okay {
		// Voltha-Protos 3.2.2 and below
		return time.Unix(int64(tsFloat), 0), nil
	}
	tsInt64, okay := tsIntf.(int64)
	if okay {
		if tsInt64 > 10000000000000 {
			// sometimes it's in nanoseconds
			return time.Unix(tsInt64/1000000000, tsInt64%1000000000), nil
		} else {
			// sometimes it's in seconds
			return time.Unix(tsInt64/1000, 0), nil
		}
	}
	return time.Time{}, errors.New("Failed to decode timestamp")
}

// Extract the header, as well as a few other items that might be of interest
func DecodeHeader(b []byte, ts time.Time) (*EventHeader, error) {
	m := &voltha.Event{}
	if err := proto.Unmarshal(b, m); err != nil {
		return nil, err
	}

	header := m.Header
	cat := voltha.EventCategory_Types_name[int32(header.Category)]
	subCat := voltha.EventSubCategory_Types_name[int32(header.SubCategory)]
	evType := voltha.EventType_Types_name[int32(header.Type)]
	raised, err := DecodeTimestamp(header.RaisedTs)
	if err != nil {
		return nil, err
	}
	reported, err := DecodeTimestamp(header.ReportedTs)
	if err != nil {
		return nil, err
	}

	// Opportunistically try to extract the device_id and title from a kpi_event2
	// note that there might actually be multiple_slice data, so there could
	// be multiple device_id, multiple title, etc.
	device_ids := make(map[string]interface{})
	titles := make(map[string]interface{})

	kpi := m.GetKpiEvent2()
	if kpi != nil {
		sliceList := kpi.SliceData
		if len(sliceList) > 0 {
			slice := sliceList[0]
			if slice != nil {
				metadata := slice.Metadata
				deviceId := metadata.DeviceId
				device_ids[deviceId] = slice

				title := metadata.Title
				titles[title] = slice
			}
		}
	}

	// Opportunistically try to pull a resource_id and title from a DeviceEvent
	// There can only be one resource_id and title from a DeviceEvent, so it's easier
	// than dealing with KPI_EVENT2.
	deviceEvent := m.GetDeviceEvent()
	if deviceEvent != nil {
		titles[deviceEvent.DeviceEventName] = deviceEvent
		device_ids[deviceEvent.ResourceId] = deviceEvent
	}

	device_id_keys := make([]string, len(device_ids))
	i := 0
	for k := range device_ids {
		device_id_keys[i] = k
		i++
	}

	title_keys := make([]string, len(titles))
	i = 0
	for k := range titles {
		title_keys[i] = k
		i++
	}

	evHeader := EventHeader{Category: cat,
		SubCategory: subCat,
		Type:        evType,
		Raised_ts:   raised,
		Reported_ts: reported,
		Device_ids:  device_id_keys,
		Timestamp:   ts,
		Titles:      title_keys}

	return &evHeader, nil
}

// Print the full message, either in JSON or in GRPCURL-human-readable format,
// depending on which grpcurl formatter is passed in.
func PrintMessage(outputAs string, b []byte) error {
	ms := &voltha.Event{}
	if err := proto.Unmarshal(b, ms); err != nil {
		return err
	}

	if outputAs == "json" {
		marshaler := jsonpb.Marshaler{EmitDefaults: true}
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

// Print just the enriched EventHeader. This is either in JSON format, or in
// table format.
func PrintEventHeader(outputAs string, outputFormat string, hdr *EventHeader) error {
	if outputAs == "json" {
		asJson, err := json.Marshal(hdr)
		if err != nil {
			return fmt.Errorf("Error marshalling JSON: %v", err)
		} else {
			fmt.Printf("%s\n", asJson)
		}
	} else {
		f := format.Format(outputFormat)
		output, err := f.ExecuteFixedWidth(DefaultWidths, false, *hdr)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", output)
	}
	return nil
}

// Start output, print any column headers or other start characters
func (options *EventListenOpts) StartOutput(outputFormat string) error {
	if options.OutputAs == "json" {
		fmt.Println("[")
	} else if (options.OutputAs == "table") && !options.ShowBody {
		f := format.Format(outputFormat)
		output, err := f.ExecuteFixedWidth(DefaultWidths, true, nil)
		if err != nil {
			return err
		}
		fmt.Println(output)
	}
	return nil
}

// Finish output, print any column footers or other end characters
func (options *EventListenOpts) FinishOutput() {
	if options.OutputAs == "json" {
		fmt.Println("]")
	}
}

func (options *EventListenOpts) Execute(args []string) error {
	ProcessGlobalOptions()
	if GlobalConfig.Current().Kafka == "" {
		return errors.New("Kafka address is not specified")
	}

	config := sarama.NewConfig()
	config.ClientID = "go-kafka-consumer"
	config.Consumer.Return.Errors = true
	config.Version = sarama.V1_0_0_0
	brokers := []string{GlobalConfig.Current().Kafka}

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return err
	}

	defer func() {
		if err := client.Close(); err != nil {
			panic(err)
		}
	}()

	consumer, consumerErrors, highwaterMarks, err := startConsumer([]string{"voltha.events"}, client)
	if err != nil {
		return err
	}

	highwater := highwaterMarks["voltha.events"]

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
		outputFormat = GetCommandOptionWithDefault("events-listen", "format", DEFAULT_EVENT_FORMAT)
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
				hdr, err := DecodeHeader(msg.Value, msg.Timestamp)
				if err != nil {
					log.Printf("Error decoding header %v\n", err)
					continue
				}

				match := false
				if headerFilter != nil {
					var err error
					if match, err = headerFilter.Evaluate(*hdr); err != nil {
						log.Printf("%v\n", err)
					}
				} else {
					match = true
				}

				if !match {
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
						if err := PrintMessage(options.OutputAs, msg.Value); err != nil {
							log.Printf("%v\n", err)
						}
					} else {
						if err := PrintEventHeader(options.OutputAs, outputFormat, hdr); err != nil {
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
func startConsumer(topics []string, client sarama.Client) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError, map[string]int64, error) {
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

		hw, err := client.GetOffset("voltha.events", partitions[0], sarama.OffsetNewest)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Error in consume() getting highwater: Topic %v Partitions: %v (%s)", topic, partitions, err)
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
