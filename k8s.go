package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"log"
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
	clientset *kubernetes.Clientset
}

func NewLogStreamer(clientset *kubernetes.Clientset) *LogStreamer {
	return &LogStreamer{
		clientset: clientset,
	}
}

func (s *LogStreamer) Stream(src *LogSource) (*LogStream, error) {
	log.Println("creating a new stream for", src)

	since := int64(1)
	req := s.clientset.CoreV1().Pods(src.Namespace).GetLogs(src.Pod, &v1.PodLogOptions{
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

	go func() {
		<-stop
		log.Println("stopping stream", src)
		cancel()
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
