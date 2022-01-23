package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"sync"
)

type LogSource struct {
	Namespace string
	Pod       string
	Container string
}

func (l LogSource) String() string {
	return fmt.Sprintf("%s/%s/%s", l.Namespace, l.Pod, l.Container)
}

type LogStreamer struct {
	mu        sync.RWMutex
	clientset *kubernetes.Clientset
	streams   map[*LogSource]*LogStream
}

func NewLogStreamer(clientset *kubernetes.Clientset) *LogStreamer {
	return &LogStreamer{clientset: clientset, streams: make(map[*LogSource]*LogStream)}
}

func (k *LogStreamer) Stream(src *LogSource) (*LogStream, error) {
	k.mu.RLock()
	if stream, ok := k.streams[src]; ok {
		log.Println("found existing stream for", src)
		return stream, nil
	}
	k.mu.RUnlock()

	log.Println("creating a new stream for", src)

	since := int64(1)
	req := k.clientset.CoreV1().Pods(src.Namespace).GetLogs(src.Pod, &v1.PodLogOptions{
		Container:    src.Container,
		Follow:       true,
		SinceSeconds: &since,
	})
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := req.Stream(ctx)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "stream logs")
	}

	stop := make(chan bool, 1)
	logs := make(chan string, 1)
	logStream := &LogStream{
		logs:   logs,
		stop:   stop,
		source: src,
	}
	k.mu.Lock()
	k.streams[src] = logStream
	k.mu.Unlock()

	go func() {
		<-stop
		log.Println("stopping stream", src)
		cancel()

		k.mu.Lock()
		delete(k.streams, src)
		k.mu.Unlock()
	}()

	scanner := bufio.NewScanner(stream)
	go func() {
		defer close(logs)
		defer stream.Close()

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				log.Println("cancelled")
				return
			case logs <- scanner.Text():
			}
		}
	}()

	return logStream, nil
}
