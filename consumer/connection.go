package main

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"time"
)

type Consumer[T comparable] struct {
	reader *kafka.Reader
	dialer *kafka.Dialer
	topic  string
}

// CreateConnection Connects to Kafka and returns us the reader object
func (c *Consumer[T]) CreateConnection() {
	c.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     c.topic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
		MaxWait:   time.Millisecond * 10,
		Dialer:    c.dialer,
	})
	c.reader.SetOffset(0)
}

// Read model — our tables and callback — we pass another table process to here for merge tables
// We are looking for new messages until the context is over from Kafka.
// The context will be killed after 80 milliseconds when any message does not receive.
func (c *Consumer[T]) Read(model T, callback func(T, error)) {
	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*80)
		message, err := c.reader.ReadMessage(ctx)

		if err != nil {
			callback(model, err)
			return
		}
		err = json.Unmarshal(message.Value, &model)
		if err != nil {
			callback(model, err)
			continue
		}
		callback(model, nil)
	}
}
