package signaler

import (
	"os"
	"testing"
)

func TestSignaler(t *testing.T) {
	// Set up test environment
	os.Setenv("MAIN_CONTAINER_PID", "1") // Use PID 1, which always exists

	s := NewSignaler("main-app", "SIGHUP")

	// Test signaling
	err := s.Signal()
	if err != nil {
		// We expect an error here because we don't have permission to signal PID 1
		if err.Error() != "operation not permitted" {
			t.Fatalf("unexpected error: %v", err)
		}
	} else {
		t.Fatalf("expected 'operation not permitted' error, but got none")
	}

	// Test with invalid PID
	os.Setenv("MAIN_CONTAINER_PID", "invalid")
	err = s.Signal()
	if err == nil {
		t.Fatalf("expected error with invalid PID, but got none")
	}
}
