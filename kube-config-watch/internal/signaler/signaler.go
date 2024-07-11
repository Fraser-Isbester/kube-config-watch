package signaler

import (
	"fmt"
	"os"
	"syscall"
)

type Signaler struct {
	mainContainerName string
	signalType        string
}

func NewSignaler(mainContainerName, signalType string) *Signaler {
	return &Signaler{
		mainContainerName: mainContainerName,
		signalType:        signalType,
	}
}

func (s *Signaler) Signal() error {
	pid, err := s.findMainContainerPID()
	if err != nil {
		return err
	}

	var sig syscall.Signal
	switch s.signalType {
	case "SIGHUP":
		sig = syscall.SIGHUP
	case "SIGTERM":
		sig = syscall.SIGTERM
	default:
		return fmt.Errorf("unsupported signal type: %s", s.signalType)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(sig)
}

func (s *Signaler) findMainContainerPID() (int, error) {
	// Implementation depends on how the main container's PID is made available
	// For example, it could be passed as an environment variable or written to a shared file
	// This is a simplified example:
	pidStr := os.Getenv("MAIN_CONTAINER_PID")
	if pidStr == "" {
		return 0, fmt.Errorf("main container PID not found")
	}

	var pid int
	_, err := fmt.Sscanf(pidStr, "%d", &pid)
	if err != nil {
		return 0, err
	}

	return pid, nil
}
