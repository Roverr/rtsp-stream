package streaming

import (
	"os/exec"
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
