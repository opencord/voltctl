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
	"errors"
	"fmt"
	// kafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/Shopify/sarama"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/opencord/voltctl/pkg/model"
	"os"
	"os/signal"
	"strings"
)

type EventDumpOpts struct {
	ListOutputOptions
}

type EventOpts struct {
	EventDump EventDumpOpts `command:"dump"`
}

var eventOpts = EventOpts{}

type EventHeader struct {
	category    string
	subCategory string
	evType      string
}

func RegisterEventCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("event", "event commands", "Commands for dumping events", &eventOpts)
	if err != nil {
		Error.Fatalf("Unable to register event commands with voltctl command parser: %s", err.Error())
	}
}

func DecodeHeader(md *desc.MessageDescriptor, b []byte) (*EventHeader, error) {
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

	evHeader := EventHeader{category: model.GetEnumString(header, "category", cat),
		subCategory: model.GetEnumString(header, "sub_category", subCat),
		evType:      model.GetEnumString(header, "type", evType)}

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

	consumer, errors := consume([]string{"voltha.events"}, master)

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

	// Get signnal for finish
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-consumer:
				msgCount++
				//hdr, err := DecodeHeader(eventMessage, msg.Value)
				PrintMessage(formatter, eventMessage, msg.Value)
			case consumerError := <-errors:
				msgCount++
				fmt.Println("Received consumerError ", string(consumerError.Topic), string(consumerError.Partition), consumerError.Err)
				doneCh <- struct{}{}
			case <-signals:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	fmt.Println("Processed", msgCount, "messages")

	return nil
}

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
			fmt.Printf("Topic %v Partitions: %v", topic, partitions)
			panic(err)
		}
		fmt.Println(" Start consuming topic ", topic)
		go func(topic string, consumer sarama.PartitionConsumer) {
			for {
				select {
				case consumerError := <-consumer.Errors():
					errors <- consumerError
					//fmt.Println("consumerError: ", consumerError.Err)

				case msg := <-consumer.Messages():
					consumers <- msg
					//fmt.Println("Got message on topic ", topic)
				}
			}
		}(topic, consumer)
	}

	return consumers, errors
}
