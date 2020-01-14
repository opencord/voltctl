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
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/opencord/voltctl/pkg/filter"
	"github.com/opencord/voltctl/pkg/model"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

type EventDumpOpts struct {
	Format      string `long:"format" value-name:"FORMAT" default:"" description:"Format to use to output structured data"`
	Filter      string `short:"f" long:"filter" default:"" value-name:"FILTER" description:"Only display results that match filter"`
	OnlyHeaders bool   `short:"d" long:"only-headers" description:"Displays headers but not bodies of events"`
	Count       int    `short:"c" long:"count" default:"-1" value-name:"LIMIT" description:"Limit the count of messages that will be printed"`
	Now         bool   `short:"n" long:"now" description:"Stop printing events when current time is reached"`
	Timeout     int    `short:"t" long:"idle" default:"900" value-name:"SECONDS" description:"Timeout if no message received within specified seconds"`
}

type EventOpts struct {
	EventDump EventDumpOpts `command:"dump"`
}

var eventOpts = EventOpts{}

type EventHeader struct {
	Category    string   `json:"category"`
	SubCategory string   `json:"sub_category"`
	EvType      string   `json:"type"`
	Raised_ts   float32  `json:"raised_ts"`
	Reported_ts float32  `json:"reported_ts"`
	Device_ids  []string `json:"device_ids"` // Opportunistically collected list of device_ids
	Titles      []string `json:"titles"`     // Opportunistically collected list of titles
	Timestamp   int64    `json:"timestamp"`  // Timestamp from Kafka
}

func RegisterEventCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("event", "event commands", "Commands for dumping events", &eventOpts)
	if err != nil {
		Error.Fatalf("Unable to register event commands with voltctl command parser: %s", err.Error())
	}
}

// Extract the header, as well as a few other items that might be of interest
func DecodeHeader(md *desc.MessageDescriptor, b []byte, ts time.Time) (*EventHeader, error) {
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

	catIntf, err := header.TryGetFieldByName("category")
	if err != nil {
		return nil, err
	}
	cat := catIntf.(int32)

	subCatIntf, err := header.TryGetFieldByName("sub_category")
	if err != nil {
		return nil, err
	}
	subCat := subCatIntf.(int32)

	typeIntf, err := header.TryGetFieldByName("type")
	if err != nil {
		return nil, err
	}
	evType := typeIntf.(int32)

	raisedIntf, err := header.TryGetFieldByName("raised_ts")
	if err != nil {
		return nil, err
	}
	raised := raisedIntf.(float32)

	reportedIntf, err := header.TryGetFieldByName("reported_ts")
	if err != nil {
		return nil, err
	}
	reported := reportedIntf.(float32)

	// Opportunistically try to extract the device_id and title from a kpi_event2
	// note that there might actually be multiple_slice data, so there could
	// be multiple device_id, multiple title, etc.
	device_ids := make(map[string]interface{})
	titles := make(map[string]interface{})
	kpiIntf, err := m.TryGetFieldByName("kpi_event2")
	if err == nil {
		kpi, ok := kpiIntf.(*dynamic.Message)
		if ok == true && kpi != nil {
			sliceListIntf, err := kpi.TryGetFieldByName("slice_data")
			if err == nil {
				sliceIntf, ok := sliceListIntf.([]interface{})
				if ok == true && len(sliceIntf) > 0 {
					slice, ok := sliceIntf[0].(*dynamic.Message)
					if ok == true && slice != nil {
						metadataIntf, err := slice.TryGetFieldByName("metadata")
						if err == nil {
							metadata, ok := metadataIntf.(*dynamic.Message)
							if ok == true && metadata != nil {
								deviceIdIntf, err := metadataIntf.(*dynamic.Message).TryGetFieldByName("device_id")
								if err == nil {
									device_ids[deviceIdIntf.(string)] = slice
								}
								titleIntf, err := metadataIntf.(*dynamic.Message).TryGetFieldByName("title")
								if err == nil {
									titles[titleIntf.(string)] = slice
								}
							}
						}
					}
				}
			}
		}
	}

	device_id_keys := make([]string, len(device_ids))
	i := 0
	for k, _ := range device_ids {
		device_id_keys[i] = k
		i++
	}

	title_keys := make([]string, len(titles))
	i = 0
	for k, _ := range titles {
		title_keys[i] = k
		i++
	}

	evHeader := EventHeader{Category: model.GetEnumString(header, "category", cat),
		SubCategory: model.GetEnumString(header, "sub_category", subCat),
		EvType:      model.GetEnumString(header, "type", evType),
		Raised_ts:   raised,
		Reported_ts: reported,
		Device_ids:  device_id_keys,
		Timestamp:   ts.Unix(),
		Titles:      title_keys}

	return &evHeader, nil
}

func PrintMessage(f grpcurl.Formatter, md *desc.MessageDescriptor, b []byte) error {
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

func PrintHeader(format string, hdr *EventHeader) {
	if format == "json" {
		// only prints {}. weird.
		asJson, err := json.Marshal(hdr)
		if err != nil {
			log.Printf("Error marshalling JSON: %v", err)
		} else {
			fmt.Printf("%s\n", asJson)
		}
	} else {
		// Prints golang
		fmt.Printf("%v\n", *hdr)
	}
}

func GetEventMessageDesc() (*desc.MessageDescriptor, error) {
	// This is a very long-winded way to get a message descriptor

	// any descriptor on the file will do
	descriptor, _, err := GetMethod("update-log-level")
	if err != nil {
		return nil, err
	}

	// get the symbol for voltha.events
	eventSymbol, err := descriptor.FindSymbol("voltha.Event")
	if err != nil {
		return nil, err
	}

	/*
	 * EventSymbol is a Descriptor, but not a MessageDescrptior,
	 * so we can't look at it's fields yet. Go back to the file,
	 * call FindMessage to get the Message, then ...
	 */

	eventFile := eventSymbol.GetFile()
	eventMessage := eventFile.FindMessage("voltha.Event")

	return eventMessage, nil
}

func (options *EventDumpOpts) Execute(args []string) error {
	ProcessGlobalOptions()
	if GlobalConfig.Kafka == "" {
		return errors.New("Kafka address is not specified")
	}

	eventMessage, err := GetEventMessageDesc()
	if err != nil {
		return err
	}

	config := sarama.NewConfig()
	config.ClientID = "go-kafka-consumer"
	config.Consumer.Return.Errors = true
	config.Version = sarama.V1_0_0_0
	brokers := []string{GlobalConfig.Kafka}
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return err
	}

	defer func() {
		if err := master.Close(); err != nil {
			panic(err)
		}
	}()

	consumer, consumerErrors := consume([]string{"voltha.events"}, master)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// Count how many message processed
	msgCount := 0

	var formatter grpcurl.Formatter
	if options.Format == "json" {
		// need a descriptor source, any method will do
		descriptor, _, _ := GetMethod("device-list")
		if err != nil {
			return err
		}
		formatter = grpcurl.NewJSONFormatter(false, grpcurl.AnyResolverFromDescriptorSource(descriptor))
	} else {
		formatter = grpcurl.NewTextFormatter(false)
	}

	var headerFilter *filter.Filter
	if options.Filter != "" {
		headerFilterVal, err := filter.Parse(options.Filter)
		if err != nil {
			return fmt.Errorf("Failed to parse filter: %v", err)
		}
		headerFilter = &headerFilterVal
	}

	// Get signnal for finish
	doneCh := make(chan struct{})
	go func() {
		if options.Format == "json" {
			fmt.Println("[")
		}
		// Count how many messages printed
		count := 0
	Loop:
		for {
			// Initialize the idle timeout timer
			timeoutTimer := time.NewTimer(time.Duration(options.Timeout) * time.Second)
			select {
			case msg := <-consumer:
				msgCount++
				hdr, err := DecodeHeader(eventMessage, msg.Value, msg.Timestamp)
				if err != nil {
					log.Printf("Error decoding header %v\n", err)
					continue
				}
				if headerFilter != nil && !headerFilter.Evaluate(*hdr) {
					continue
				}
				if count > 0 {
					if options.Format == "json" {
						fmt.Println(",")
					}
				}
				if options.OnlyHeaders {
					PrintHeader(options.Format, hdr)
				} else {
					PrintMessage(formatter, eventMessage, msg.Value)
				}
				count++
				if (options.Count > 0) && (count >= options.Count) {
					doneCh <- struct{}{}
					break Loop
				}
				if (options.Now) && (hdr.Timestamp >= time.Now().Unix()) {
					doneCh <- struct{}{}
					break Loop
				}
				if !timeoutTimer.Stop() {
					<-timeoutTimer.C
				}
			case consumerError := <-consumerErrors:
				msgCount++
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

	if options.Format == "json" {
		fmt.Println("]")
	}

	log.Printf("Received %d messages.", msgCount)

	return nil
}

// Consume message from Sarama and send them out on a channel.
// Supports multiple topics.
// Taken from Sarama example consumer.
func consume(topics []string, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError) {
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)
	for _, topic := range topics {
		if strings.Contains(topic, "__consumer_offsets") {
			continue
		}
		partitions, _ := master.Partitions(topic)
		// this only consumes partition no 1, you would probably want to consume all partitions
		consumer, err := master.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
		if nil != err {
			log.Printf("Error in consume(): Topic %v Partitions: %v", topic, partitions)
			panic(err)
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

	return consumers, errors
}
