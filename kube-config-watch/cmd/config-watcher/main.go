package main

import (
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/fraser-isbester/kube-config-watch/internal/signaler"
	"github.com/fraser-isbester/kube-config-watch/internal/watcher"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Fatal(err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	namespace := os.Getenv("NAMESPACE")
	resourceType := os.Getenv("RESOURCE_TYPE")
	resourceName := os.Getenv("RESOURCE_NAME")
	mainContainerName := os.Getenv("MAIN_CONTAINER_NAME")
	signalType := os.Getenv("SIGNAL_TYPE")

	w := watcher.NewWatcher(clientset, namespace, resourceType, resourceName)
	s := signaler.NewSignaler(mainContainerName, signalType)

	log.Println("Starting config watcher...")
	for {
		if changed, err := w.Watch(); err != nil {
			log.Printf("Error watching resource: %v", err)
		} else if changed {
			if err := s.Signal(); err != nil {
				log.Printf("Error signaling main container: %v", err)
			} else {
				log.Println("Successfully signaled main container")
			}
		}
	}
}
