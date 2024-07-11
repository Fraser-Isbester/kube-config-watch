package watcher

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// K8sClient is an interface that abstracts the methods we need from the Kubernetes clientset
type K8sClient interface {
	CoreV1() corev1.CoreV1Interface
}

type Watcher struct {
	client       K8sClient
	namespace    string
	resourceType string
	resourceName string
	lastVersion  string
}

func NewWatcher(client K8sClient, namespace, resourceType, resourceName string) *Watcher {
	return &Watcher{
		client:       client,
		namespace:    namespace,
		resourceType: resourceType,
		resourceName: resourceName,
	}
}

func (w *Watcher) Watch() (bool, error) {
	var resourceVersion string

	switch w.resourceType {
	case "secret":
		secret, err := w.client.CoreV1().Secrets(w.namespace).Get(context.TODO(), w.resourceName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		resourceVersion = secret.ResourceVersion
	case "configmap":
		configMap, err := w.client.CoreV1().ConfigMaps(w.namespace).Get(context.TODO(), w.resourceName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		resourceVersion = configMap.ResourceVersion
	default:
		return false, fmt.Errorf("unsupported resource type: %s", w.resourceType)
	}

	if resourceVersion != w.lastVersion {
		w.lastVersion = resourceVersion
		return true, nil
	}

	return false, nil
}
