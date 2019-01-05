package streaming

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/Roverr/hotstreak"
)

// Stream describes a given host's streaming
type Stream struct {
	CMD         *exec.Cmd            `json:"-"`
	Mux         *sync.Mutex          `json:"-"`
	Path        string               `json:"path"`
	Streak      *hotstreak.Hotstreak `json:"-"`
	OriginalURI string               `json:"-"`
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
