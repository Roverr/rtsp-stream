package streaming

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/sirupsen/logrus"
)

// IProcessor is an interface describing a processor service
type IProcessor interface {
	NewProcess(path, URI string) *exec.Cmd
	Restart(stream *Stream, path string) error
}

// Processor is the main type for creating new processes
type Processor struct {
	keepFiles   bool
	audio       bool
	loggingOpts config.ProcessLogging
}

// Type check
var _ IProcessor = (*Processor)(nil)

// NewProcessor creates a new instance of a processor
func NewProcessor(
	keepFiles bool,
	audio bool,
	loggingOpts config.ProcessLogging,
) *Processor {
	return &Processor{audio, keepFiles, loggingOpts}
}

// getHLSFlags are for getting the flags based on the config context
func (p Processor) getHLSFlags() string {
	if p.keepFiles {
		return "append_list"
	}
	return "delete_segments+append_list"
}

// NewProcess creates only the process for the stream
func (p Processor) NewProcess(path, URI string) *exec.Cmd {
	os.MkdirAll(path, os.ModePerm)
	processCommands := []string{
		"-y",
		"-fflags",
		"nobuffer",
		"-rtsp_transport",
		"tcp",
		"-i",
		URI,
		"-vsync",
		"0",
		"-copyts",
		"-vcodec",
		"copy",
		"-movflags",
		"frag_keyframe+empty_moov",
	}
	if p.audio {
		processCommands = append(processCommands, "-an")
	}
	processCommands = append(processCommands,
		"-hls_flags",
		p.getHLSFlags(),
		"-f",
		"hls",
		"-segment_list_flags",
		"live",
		"-hls_time",
		"1",
		"-hls_list_size",
		"3",
		"-hls_segment_filename",
		fmt.Sprintf("%s/%%d.ts", path),
		fmt.Sprintf("%s/index.m3u8", path),
	)
	cmd := exec.Command("ffmpeg", processCommands...)
	return cmd
}

// Restart uses the processor to restart a given stream
func (p Processor) Restart(stream *Stream, path string) error {
	stream.Mux.Lock()
	defer stream.Mux.Unlock()
	stream.CMD = p.NewProcess(stream.StorePath, stream.OriginalURI)
	if p.loggingOpts.Enabled {
		stream.CMD.Stderr = stream.Logger
		stream.CMD.Stdout = stream.Logger
	}
	stream.Streak.Activate()
	go func() {
		if err := stream.CMD.Run(); err != nil {
			logrus.Error(err)
		}
	}()
	logrus.Infof("%s has been restarted", path)
	return nil
}
