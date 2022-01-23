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
	src := &LogSource{"dev", "pyapp-6d76fbb595-jdb8h", "python"}

	stream, err := streamer.Stream(src)
	if err != nil {
		log.Fatal(err)
	}

	broker := NewStreamBroker(stream)

	c1 := broker.Subscribe()
	c2 := broker.Subscribe()

	go func() {
		for {
			log.Println(<-c1.messages)
		}
	}()

	go func() {
		for {
			log.Println(<-c2.messages)
		}
	}()

	go func() {
		<-time.After(time.Second * 20)
		broker.Unsubscribe(c1)
		broker.Unsubscribe(c2)
	}()

	broker.Dispatch()
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
