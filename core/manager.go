package core

import (
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
type Manager struct{}

// Type check
var _ IManager = (*Manager)(nil)

// Start is to manage the start of the transcoding
func (m Manager) Start(cmd *exec.Cmd, physicalPath string) chan bool {
	var once sync.Once
	streamResolved := make(chan bool, 1)
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
	go func() {
		if err := cmd.Run(); err != nil {
			logrus.Error(err)
			once.Do(func() { streamResolved <- false })
		}
	}()

	return streamResolved
}
