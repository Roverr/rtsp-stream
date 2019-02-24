package streaming

import (
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/Roverr/hotstreak"
)

// Stream describes a given host's streaming
type Stream struct {
	CMD         *exec.Cmd            `json:"-"`
	Mux         *sync.RWMutex        `json:"-"`
	Path        string               `json:"path"`
	Streak      *hotstreak.Hotstreak `json:"-"`
	OriginalURI string               `json:"-"`
	StorePath   string               `json:"-"`
}

// CleanProcess makes sure that the transcoding process is killed correctly
func (strm *Stream) CleanProcess() error {
	strm.Mux.Lock()
	strm.Streak.Deactivate()
	defer strm.cleanDir()
	defer strm.Mux.Unlock()
	if err := strm.CMD.Process.Kill(); err != nil {
		if strings.Contains(err.Error(), "process already finished") {
			return nil
		}
		if strings.Contains(err.Error(), "signal: killed") {
			return nil
		}
		return err
	}
	return nil
}

// cleanDir cleans the directory that includes files of the already stopped stream
func (strm *Stream) cleanDir() {
	logrus.Debugf("%s directory is being cleaned", strm.StorePath)
	if err := os.RemoveAll(strm.StorePath); err != nil {
		logrus.Error(err)
	}
}
