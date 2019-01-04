package streaming

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/Roverr/hotstreak"
	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/sirupsen/logrus"
)

// Stream describes a given host's streaming
type Stream struct {
	CMD         *exec.Cmd            `json:"-"`
	Mux         *sync.Mutex          `json:"-"`
	Path        string               `json:"path"`
	Streak      *hotstreak.Hotstreak `json:"-"`
	OriginalURI string               `json:"-"`
}

// Restart is a function to restart a given stream's transcoding process
func (strm *Stream) Restart(spec *config.Specification, path string) error {
	strm.Mux.Lock()
	defer strm.Mux.Unlock()
	strm.CMD, _, _ = NewProcess(strm.OriginalURI, spec)
	strm.Streak.Activate()
	go func() {
		logrus.Infof("%s has been restarted", path)
		err := strm.CMD.Run()
		if err != nil {
			logrus.Error(err)
		}
	}()
	return nil
}

// CleanProcess makes sure that the transcoding process is killed correctly
func (strm *Stream) CleanProcess() error {
	strm.Mux.Lock()
	strm.Streak.Deactivate()
	defer strm.Mux.Unlock()
	if err := strm.CMD.Process.Kill(); err != nil {
		if strings.Contains(err.Error(), "process already finished") {
			return nil
		}
		return err
	}
	return nil
}
