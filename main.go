package main

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"time"
)

func main() {
	kubeconfigPath := "/Users/abdus/Desktop/kubeconfig-kube.yml"
	clientset, err := NewKubernetesClient(kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}

	streamer := NewLogStreamer(clientset)

	broker := NewLogBroker(streamer)
	for i := 0; i < 2; i++ {
		container := "python"
		if i%2 == 1 {
			container = "client"
		}

		src := &LogSource{"dev", "pyapp-6d76fbb595-jdb8h", container}

		consumer, err := broker.Subscribe(src)
		if err != nil {
			log.Fatal(err)
		}
		go func(c *LogConsumer) {
			defer broker.Unsubscribe(c)

			for {
				select {
				case <-time.After(time.Second * 5):
					log.Println("timeout")
					return
				case m := <-c.messages:
					log.Println(m)
				}
			}
		}(consumer)
	}

	broker.PumpMessages()
}

func NewKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "read kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "create kubernetes client")
	}
	return clientset, nil
}
