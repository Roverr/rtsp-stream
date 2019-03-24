package streaming

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/Roverr/hotstreak"
	"github.com/kennygrant/sanitize"
	"github.com/sirupsen/logrus"
)

// ErrInvalidHost describes an error for a hostname that is considered invalid if it's empty
var ErrInvalidHost = errors.New("Invalid hostname")

// ErrUnparsedURL describes an error that occours when the parsing process cannot be deemed as successful
var ErrUnparsedURL = errors.New("URL is not parsed correctly")

// IProcessor is an interface describing a processor service
type IProcessor interface {
	NewProcess(URI string) *exec.Cmd
	NewStream(URI string) (*Stream, string)
	Restart(stream *Stream, path string) error
}

// Processor is the main type for creating new processes
type Processor struct {
	storeDir string
}

// Type check
var _ IProcessor = (*Processor)(nil)

// NewProcessor creates a new instance of a processor
func NewProcessor(storeDir string) *Processor {
	return &Processor{storeDir}
}

// NewProcess creates only the process for the stream
func (p Processor) NewProcess(URI string) *exec.Cmd {
	dirPath, newPath, err := createDirectoryForURI(URI, p.storeDir)
	if err != nil {
		logrus.Error("Error happened while getting directory name in creating process", dirPath)
		return nil
	}

	cmd := exec.Command(
		"ffmpeg",
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
		"-an",
		"-hls_flags",
		"delete_segments+append_list",
		"-f",
		"hls",
		"-segment_list_flags",
		"live",
		"-hls_time",
		"1",
		"-hls_list_size",
		"3",
		"-hls_segment_filename",
		fmt.Sprintf("%s/%%d.ts", newPath),
		fmt.Sprintf("%s/index.m3u8", newPath),
	)
	return cmd
}

// NewStream creates a new transcoding process for ffmpeg
func (p Processor) NewStream(URI string) (*Stream, string) {
	dirPath, newPath, err := createDirectoryForURI(URI, p.storeDir)
	if err != nil {
		logrus.Error("Error happened while getting directory name", dirPath)
		return nil, ""
	}
	cmd := p.NewProcess(URI)
	stream := Stream{
		CMD:       cmd,
		Mux:       &sync.RWMutex{},
		Path:      fmt.Sprintf("/%s/index.m3u8", filepath.Join("stream", dirPath)),
		StorePath: newPath,
		Streak: hotstreak.New(hotstreak.Config{
			Limit:      10,
			HotWait:    time.Minute * 2,
			ActiveWait: time.Minute * 4,
		}).Activate(),
		OriginalURI: URI,
	}
	logrus.Debugf("Created stream with storepath %s", stream.StorePath)
	return &stream, fmt.Sprintf("%s/index.m3u8", newPath)
}

// Restart uses the processor to restart a given stream
func (p Processor) Restart(strm *Stream, path string) error {
	strm.Mux.Lock()
	defer strm.Mux.Unlock()
	strm.CMD = p.NewProcess(strm.OriginalURI)
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

// ValidateURL checks if everything is present for the given URL
func ValidateURL(URL *url.URL) error {
	if URL == nil {
		return ErrUnparsedURL
	}
	if URL.Hostname() == "" {
		return ErrInvalidHost
	}
	return nil
}

// GetURIDirectory is a function to create a directory string from an URI
func GetURIDirectory(URI string) (string, error) {
	URL, err := url.Parse(URI)
	if err != nil {
		return "", err
	}
	if err = ValidateURL(URL); err != nil {
		return "", err
	}
	return sanitize.BaseName(fmt.Sprintf("%s-%s", URL.Hostname(), sanitize.Path(URL.Path))), nil
}

// createDirectoryForURI is to create a safe path based on the received URI
func createDirectoryForURI(URI, storeDir string) (dirPath, newPath string, err error) {
	dirPath, err = GetURIDirectory(URI)
	if err != nil {
		return
	}

	newPath = fmt.Sprintf("%s/%s", storeDir, dirPath)
	err = os.MkdirAll(newPath, os.ModePerm)
	return
}
