package watcher

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestWatcher(t *testing.T) {
	// Create a fake clientset
	clientset := fake.NewSimpleClientset()

	// Create a test secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-secret",
			Namespace:       "default",
			ResourceVersion: "1",
		},
	}
	_, err := clientset.CoreV1().Secrets("default").Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("error creating test secret: %v", err)
	}

	// Create a watcher
	w := NewWatcher(clientset, "default", "secret", "test-secret")

	// Test initial watch
	changed, err := w.Watch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Errorf("expected change on initial watch")
	}

	// Test second watch (no change)
	changed, err = w.Watch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Errorf("expected no change on second watch")
	}

	// Update the secret
	secret.ResourceVersion = "2"
	_, err = clientset.CoreV1().Secrets("default").Update(context.TODO(), secret, metav1.UpdateOptions{})
	if err != nil {
		t.Fatalf("error updating test secret: %v", err)
	}

	// Test watch after update
	changed, err = w.Watch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Errorf("expected change after update")
	}
}
