package main

import (
	"log"
	"strconv"
	"sync"
)

type StreamConsumer struct {
	id       string
	source   *LogSource
	messages chan string
}

type LogStream struct {
	source *LogSource
	stop   chan bool
	logs   <-chan string
}

type StreamBroker struct {
	mu        sync.Mutex
	stream    *LogStream
	listeners map[*StreamConsumer]struct{}
	closed    bool
}

func NewStreamBroker(stream *LogStream) *StreamBroker {
	return &StreamBroker{
		stream:    stream,
		listeners: make(map[*StreamConsumer]struct{}),
	}
}

func (b *StreamBroker) Subscribe() *StreamConsumer {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	listener := &StreamConsumer{
		id:       b.stream.source.String() + strconv.Itoa(len(b.listeners)),
		source:   b.stream.source,
		messages: make(chan string),
	}

	b.listeners[listener] = struct{}{}
	return listener
}

func (b *StreamBroker) Unsubscribe(listener *StreamConsumer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	log.Println("unsubscribing listener", listener.id)

	delete(b.listeners, listener)

	if len(b.listeners) == 0 {
		b.stream.stop <- true
		b.closed = true
	}
}

func (b *StreamBroker) Dispatch() {
	for m := range b.stream.logs {
		for listener := range b.listeners {
			listener.messages <- m
		}
	}
}
