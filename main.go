package main

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"strconv"
)

func main() {
	var clientset *kubernetes.Clientset
	var err error
	kubeconfigPath := os.Getenv("KUBECONFIG_PATH")
	if kubeconfigPath == "" {
		log.Println("using in-cluster config")
		clientset, err = NewKubernetesInClusterClient()
	} else {
		log.Println("using kubeconfig at", kubeconfigPath)
		clientset, err = NewKubernetesLocalClient(kubeconfigPath)
	}
	if err != nil {
		log.Fatal(err)
	}

	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		portEnv = "8080"
	}
	port, err := strconv.Atoi(portEnv)
	if err != nil {
		log.Fatal(errors.Wrap(err, "port must be numeric"))
	}

	streamer := NewLogStreamer(clientset)
	broker := NewLogBroker(streamer)

	app := NewApp(port, broker)
	log.Fatalln(app.Serve())
}

func NewKubernetesLocalClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
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

func NewKubernetesInClusterClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "read in-cluster config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "create kubernetes client")
	}
	return clientset, nil
}
