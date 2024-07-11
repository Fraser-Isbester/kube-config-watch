package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/fraser-isbester/kube-config-watch/internal/signaler"
	"github.com/fraser-isbester/kube-config-watch/internal/watcher"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
)

var kindClusterName = "test-cluster"
var kindClient *cluster.Provider
var clientset *kubernetes.Clientset

func TestMain(m *testing.M) {
	// Set up Kind cluster
	kindClient = cluster.NewProvider()
	err := kindClient.Create(kindClusterName)
	fmt.Printf("Creating Kind cluster %s...", kindClusterName)
	if err != nil {
		fmt.Printf("Failed to create Kind cluster: %v\n", err)
		os.Exit(1)
	}

	// Get the kubeconfig
	kubeconfig, err := kindClient.KubeConfig(kindClusterName, false)
	if err != nil {
		fmt.Printf("Failed to get kubeconfig: %v\n", err)
		kindClient.Delete(kindClusterName, "")
		os.Exit(1)
	}

	// Create the clientset
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		fmt.Printf("Failed to create config: %v\n", err)
		kindClient.Delete(kindClusterName, "")
		os.Exit(1)
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to create clientset: %v\n", err)
		kindClient.Delete(kindClusterName, "")
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Clean up
	defer kindClient.Delete(kindClusterName, "")
	os.Exit(code)
}

func TestWatcherWithRealCluster(t *testing.T) {
	ctx := context.Background()
	namespace := "default"
	secretName := "test-secret"
	configMapName := "test-configmap"

	// Test Secret Watcher
	t.Run("SecretWatcher", func(t *testing.T) {
		// Create a secret
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: secretName,
			},
			StringData: map[string]string{
				"key": "value",
			},
		}
		_, err := clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create secret: %v", err)
		}

		w := watcher.NewWatcher(clientset, namespace, "secret", secretName)

		// Initial watch should detect the secret
		changed, err := w.Watch()
		if err != nil || !changed {
			t.Fatalf("Expected initial watch to detect change, got changed=%v, err=%v", changed, err)
		}

		// Second watch should not detect change
		changed, err = w.Watch()
		if err != nil || changed {
			t.Fatalf("Expected second watch to not detect change, got changed=%v, err=%v", changed, err)
		}

		// Update the secret
		secret.StringData["key"] = "new-value"
		_, err = clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			t.Fatalf("Failed to update secret: %v", err)
		}

		// Watch should detect the change
		changed, err = w.Watch()
		if err != nil || !changed {
			t.Fatalf("Expected watch to detect change after update, got changed=%v, err=%v", changed, err)
		}
	})

	// Test ConfigMap Watcher
	t.Run("ConfigMapWatcher", func(t *testing.T) {
		// Create a configmap
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: configMapName,
			},
			Data: map[string]string{
				"key": "value",
			},
		}
		_, err := clientset.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("Failed to create configmap: %v", err)
		}

		w := watcher.NewWatcher(clientset, namespace, "configmap", configMapName)

		// Initial watch should detect the configmap
		changed, err := w.Watch()
		if err != nil || !changed {
			t.Fatalf("Expected initial watch to detect change, got changed=%v, err=%v", changed, err)
		}

		// Second watch should not detect change
		changed, err = w.Watch()
		if err != nil || changed {
			t.Fatalf("Expected second watch to not detect change, got changed=%v, err=%v", changed, err)
		}

		// Update the configmap
		configMap.Data["key"] = "new-value"
		_, err = clientset.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
		if err != nil {
			t.Fatalf("Failed to update configmap: %v", err)
		}

		// Watch should detect the change
		changed, err = w.Watch()
		if err != nil || !changed {
			t.Fatalf("Expected watch to detect change after update, got changed=%v, err=%v", changed, err)
		}
	})
}

func TestSignaler(t *testing.T) {
	// This test is a bit tricky as we can't easily test signaling in a container environment
	// We'll just test that the Signaler can be created without error
	s := signaler.NewSignaler("test-container", "SIGHUP")
	if s == nil {
		t.Fatal("Failed to create Signaler")
	}
}
