package main

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"strconv"
)

func main() {
	kubeconfigPath := "/Users/abdus/Desktop/kubeconfig-kube.yml"
	clientset, err := NewKubernetesClient(kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}
	streamer := NewLogStreamer(clientset)
	broker := NewLogBroker(streamer)
	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		portEnv = "8080"
	}
	port, _ := strconv.Atoi(portEnv)

	app := NewApp(port, broker)
	log.Fatalln(app.Serve())
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
