package streaming

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"

	"github.com/Roverr/hotstreak"
	"github.com/Roverr/rtsp-stream/core/config"
)

// IStream is almost like icecream, that's why it is perfect
type IStream interface {
	Start() *sync.WaitGroup
	Restart() *sync.WaitGroup
	Stop() error
}

// Stream describes a given host's streaming
type Stream struct {
	ID          string                 `json:"id"`
	Path        string                 `json:"path"`
	Running     bool                   `json:"running"`
	CMD         *exec.Cmd              `json:"-"`
	Process     IProcess               `json:"-"`
	Mux         *sync.Mutex            `json:"-"`
	Streak      *hotstreak.Hotstreak   `json:"-"`
	OriginalURI string                 `json:"-"`
	StorePath   string                 `json:"-"`
	KeepFiles   bool                   `json:"-"`
	LoggingOpts *config.ProcessLogging `json:"-"`
	Logger      *lumberjack.Logger     `json:"-"`
	WaitTimeOut time.Duration          `json:"-"`
}

// Type check
var _ IStream = (*Stream)(nil)

// NewStream creates a new transcoding process for ffmpeg
func NewStream(
	URI string,
	storingDirectory string,
	keepFiles bool,
	audio bool,
	loggingOpts config.ProcessLogging,
	waitTimeOut time.Duration,
) (*Stream, string) {
	id := uuid.New().String()
	path := fmt.Sprintf("%s/%s", storingDirectory, id)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		logrus.Error(err)
		return nil, ""
	}
	process := NewProcess(keepFiles, audio, loggingOpts)
	cmd := process.Spawn(path, URI)

	// Create nil pointer in case logging is not enabled
	cmdLogger := (*lumberjack.Logger)(nil)
	// Create logger otherwise
	if loggingOpts.Enabled {
		cmdLogger = &lumberjack.Logger{
			Filename:   fmt.Sprintf("%s/%s.log", loggingOpts.Directory, id),
			MaxSize:    loggingOpts.MaxSize,
			MaxBackups: loggingOpts.MaxBackups,
			MaxAge:     loggingOpts.MaxAge,
			Compress:   loggingOpts.Compress,
		}
		cmd.Stderr = cmdLogger
		cmd.Stdout = cmdLogger
	}
	stream := Stream{
		ID:        id,
		CMD:       cmd,
		Process:   process,
		Mux:       &sync.Mutex{},
		Path:      fmt.Sprintf("/%s/index.m3u8", filepath.Join("stream", id)),
		StorePath: path,
		Streak: hotstreak.New(hotstreak.Config{
			Limit:      10,
			HotWait:    time.Minute * 2,
			ActiveWait: time.Minute * 4,
		}).Activate(),
		OriginalURI: URI,
		KeepFiles:   keepFiles,
		LoggingOpts: &loggingOpts,
		Logger:      cmdLogger,
		Running:     false,
		WaitTimeOut: waitTimeOut,
	}
	logrus.Debugf("%s store path created | Stream", stream.StorePath)
	return &stream, id
}

// Start is starting a new transcoding process
func (strm *Stream) Start() *sync.WaitGroup {
	if strm == nil {
		return nil
	}
	strm.Mux.Lock()
	var once sync.Once
	wg := &sync.WaitGroup{}
	wg.Add(1)
	indexPath := fmt.Sprintf("%s/index.m3u8", strm.StorePath)
	// Run the transcoding, resolve stream if it errors out
	go func() {
		logrus.Debugf("%s is starting FFMPEG process | Stream", strm.ID)
		if err := strm.CMD.Run(); err != nil {
			once.Do(func() {
				logrus.Errorf("%s process could not start. | Stream\n Error: %s",
					strm.ID,
					err,
				)
				strm.Running = false
				strm.Mux.Unlock()
				wg.Done()
			})
		}
	}()
	// Try scanning for the file, resolve if we found index.m3u8
	go func() {
		for {
			_, err := os.Stat(indexPath)
			if err != nil {
				<-time.After(25 * time.Millisecond)
				continue
			}
			once.Do(func() {
				logrus.Debugf("%s - %s successfully started - index.m3u8 found | Stream",
					strm.ID,
					strm.OriginalURI,
				)
				strm.Running = true
				strm.Mux.Unlock()
				wg.Done()
			})
			return
		}
	}()
	// After a certain time if nothing happens, just error it out
	go func() {
		<-time.After(strm.WaitTimeOut)
		once.Do(func() {
			logrus.Errorf(
				"%s process starting timed out | Stream",
				strm.ID,
			)
			strm.Running = false
			strm.Mux.Unlock()
			wg.Done()
		})
	}()
	// Return channel for synchronization
	return wg
}

// Restart restarts the given CMD
func (strm *Stream) Restart() *sync.WaitGroup {
	if strm == nil {
		return nil
	}
	strm.Mux.Lock()
	if strm.CMD != nil && strm.CMD.ProcessState != nil {
		strm.CMD.Process.Kill()
	}
	strm.CMD = strm.Process.Spawn(strm.StorePath, strm.OriginalURI)
	if strm.LoggingOpts.Enabled {
		strm.CMD.Stderr = strm.Logger
		strm.CMD.Stdout = strm.Logger
	}
	strm.Streak.Activate().Hit()
	strm.Mux.Unlock()
	return strm.Start()
}

// Stop makes sure that the transcoding process is killed correctly
func (strm *Stream) Stop() error {
	strm.Mux.Lock()
	defer strm.Mux.Unlock()
	strm.Streak.Deactivate()
	strm.Running = false
	if !strm.KeepFiles {
		defer func() {
			logrus.Debugf("%s directory is being removed | Stream", strm.StorePath)
			if err := os.RemoveAll(strm.StorePath); err != nil {
				logrus.Error(err)
			}
		}()
	}
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
