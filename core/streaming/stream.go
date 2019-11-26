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

// Stream describes a given host's streaming
type Stream struct {
	Path        string                 `json:"path"`
	Running     bool                   `json:"running"`
	CMD         *exec.Cmd              `json:"-"`
	Processing  IProcessor             `json:"-"`
	Mux         *sync.RWMutex          `json:"-"`
	Streak      *hotstreak.Hotstreak   `json:"-"`
	OriginalURI string                 `json:"-"`
	StorePath   string                 `json:"-"`
	KeepFiles   bool                   `json:"-"`
	LoggingOpts *config.ProcessLogging `json:"-"`
	Logger      *lumberjack.Logger     `json:"-"`
	WaitTimeOut time.Duration          `json:"-"`
}

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
	processing := NewProcessor(keepFiles, audio, loggingOpts)
	cmd := processing.NewProcess(path, URI)

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
		CMD:       cmd,
		Mux:       &sync.RWMutex{},
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
	logrus.Debugf("Created stream with storepath %s", stream.StorePath)
	return &stream, id
}

// Start is starting a new transcoding process
func (strm *Stream) Start() *sync.WaitGroup {
	if strm == nil {
		return nil
	}
	// Start running of the process
	go func() {
		if err := strm.CMD.Run(); err != nil {
			logrus.Error(err)
		}
	}()

	// Init synchronization components
	var once sync.Once
	wg := &sync.WaitGroup{}
	wg.Add(1)

	indexPath := fmt.Sprintf("%s/index.m3u8", strm.StorePath)
	// Try scanning for the file, resolve if we found index.m3u8
	go func() {
		for {
			_, err := os.Stat(indexPath)
			if err != nil {
				<-time.After(25 * time.Millisecond)
				continue
			}
			once.Do(func() { logrus.Debugln("FASZA"); strm.Running = true; wg.Done() })
			return
		}
	}()

	// Run the transcoding, resolve stream if it errors out
	go func() {
		if err := strm.CMD.Run(); err != nil {
			once.Do(func() {
				logrus.Errorf("Error happened during starting of %s || Error: %s",
					indexPath,
					err,
				)
				strm.Running = false
				wg.Done()
			})
		}
	}()

	// After a certain time if nothing happens, just error it out
	go func() {
		<-time.After(strm.WaitTimeOut)
		once.Do(func() {
			logrus.Error(
				fmt.Errorf("%s timed out while waiting for file creation in manager start",
					indexPath,
				),
			)
			strm.Running = false
			wg.Done()
		})
	}()

	// Return channel for synchronization
	return wg
}

// Restart restarts the given CMD
func (strm *Stream) Restart() error {
	strm.Mux.Lock()
	defer strm.Mux.Unlock()
	strm.CMD = strm.Processing.NewProcess(strm.StorePath, strm.OriginalURI)
	if strm.LoggingOpts.Enabled {
		strm.CMD.Stderr = strm.Logger
		strm.CMD.Stdout = strm.Logger
	}
	strm.Streak.Activate()
	go func() {
		if err := strm.CMD.Run(); err != nil {
			logrus.Error(err)
		}
	}()
	logrus.Infof("%s has been restarted", strm.Path)
	return nil
}

// CleanProcess makes sure that the transcoding process is killed correctly
func (strm *Stream) CleanProcess() error {
	strm.Mux.Lock()
	strm.Streak.Deactivate()
	if !strm.KeepFiles {
		defer strm.cleanDir()
	}
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
