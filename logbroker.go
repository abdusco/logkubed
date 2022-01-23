package main

import (
	"log"
	"sync"
)

type LogConsumer struct {
	source   *LogSource
	messages chan string
}

type LogStream struct {
	source *LogSource
	stop   chan bool
	logs   <-chan string
}

type LogBroker struct {
	mu        sync.RWMutex
	consumers map[string]map[*LogConsumer]struct{}
	streams   map[string]*LogStream
	closed    bool
	streamer  *LogStreamer
}

func NewLogBroker(streamer *LogStreamer) *LogBroker {
	return &LogBroker{
		consumers: make(map[string]map[*LogConsumer]struct{}),
		streams:   make(map[string]*LogStream),
		streamer:  streamer,
	}
}

func (b *LogBroker) PumpMessages() {
	for {
		b.mu.RLock()
		for src, s := range b.streams {
			select {
			case m := <-s.logs:
				b.dispatch(src, m)
			default:
				continue
			}
		}
		b.mu.RUnlock()
	}
}

func (b *LogBroker) dispatch(src string, msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	if consumers, ok := b.consumers[src]; ok {
		for c := range consumers {
			c.messages <- msg
		}
	}
}

func (b *LogBroker) Subscribe(source *LogSource) (*LogConsumer, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	src := source.String()
	if _, ok := b.streams[src]; !ok {
		stream, err := b.streamer.Stream(source)
		if err != nil {
			return nil, err
		}
		b.streams[src] = stream
	}

	ch := make(chan string, 1)

	c := &LogConsumer{
		source:   source,
		messages: ch,
	}

	if _, ok := b.consumers[src]; !ok {
		b.consumers[src] = make(map[*LogConsumer]struct{})
	}
	b.consumers[src][c] = struct{}{}

	return c, nil
}

func (b *LogBroker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.closed {
		b.closed = true

		for _, s := range b.streams {
			s.stop <- true
		}

		for _, consumers := range b.consumers {
			for c := range consumers {
				close(c.messages)
			}
		}
	}
}

func (b *LogBroker) Unsubscribe(c *LogConsumer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	log.Println("unsubbing consumer", c.source)

	src := c.source.String()

	if _, ok := b.consumers[src]; ok {
		delete(b.consumers[src], c)

		if (len(b.consumers[src])) == 0 {
			b.streams[src].stop <- true
			delete(b.streams, src)
		}
	}
}
