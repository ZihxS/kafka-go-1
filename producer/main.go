package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func produce(wg *sync.WaitGroup, producer *kafka.Producer, topic string, delay time.Duration, amount int) {
	defer wg.Done()
	for i := 1; i <= amount; i++ {
		go func(topic string, i int) {
			producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Key:            []byte(fmt.Sprintf("%v-%v", topic, i)),
				Value:          []byte(fmt.Sprintf("from producer to %v: %v.", topic, i)),
			}, nil)
		}(topic, i)
		time.Sleep(delay)
	}
}

func running() {
	var wg sync.WaitGroup

	kafkaConfigMap := &kafka.ConfigMap{
		"bootstrap.servers": "localhost",
	}

	producer, err := kafka.NewProducer(kafkaConfigMap)

	if err != nil {
		panic(err)
	}

	defer producer.Close()

	go func() {
		for events := range producer.Events() {
			switch event := events.(type) {
			case *kafka.Message:
				if event.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", event.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", event.TopicPartition)
				}
			}
		}
	}()

	min, max := 1, 5

	for i := min; i <= max; i++ {
		wg.Add(1)
		go produce(&wg, producer, fmt.Sprintf("kafka-go-1-topic-%v", i), ((time.Second / 4) * time.Duration(i)), ((max+1)*50)-(min*10))
	}

	wg.Wait()
}

func main() {
	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		running()
	}()

	<-exitChan

	fmt.Println("")
	log.Println("Shutdown Signal Received!")
	log.Println("Bye Bye!")
}