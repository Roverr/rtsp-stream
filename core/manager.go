package core

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// IManager is the interface for the manager object that handles the start
// of the transcoding process
type IManager interface {
	Start(cmd *exec.Cmd, physicalPath string) chan bool
}

// Manager is describes a new object that has the start function
type Manager struct {
	timeout time.Duration
}

// Type check
var _ IManager = (*Manager)(nil)

// NewManager returns a new instance of a manager
func NewManager(timeout time.Duration) *Manager {
	return &Manager{timeout}
}

// Start is to manage the start of the transcoding
func (m Manager) Start(cmd *exec.Cmd, physicalPath string) chan bool {
	// Init synchronization components
	var once sync.Once
	streamResolved := make(chan bool, 1)

	// Try scanning for the file, resolve if we found index.m3u8
	go func() {
		for {
			_, err := os.Stat(physicalPath)
			if err != nil {
				<-time.After(25 * time.Millisecond)
				continue
			}
			once.Do(func() { streamResolved <- true })
			return
		}
	}()

	// Run the transcoding, resolve stream if it errors out
	go func() {
		if err := cmd.Run(); err != nil {
			logrus.Error(err)
			once.Do(func() { streamResolved <- false })
		}
	}()

	// After a certain time if nothing happens, just error it out
	go func() {
		<-time.After(m.timeout)
		logrus.Error(fmt.Errorf("%s timed out while waiting for file creaton in manager start", physicalPath))
		once.Do(func() { streamResolved <- false })
	}()

	// Return channel for synchronization
	return streamResolved
}
