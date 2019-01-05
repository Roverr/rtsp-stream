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
	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/kennygrant/sanitize"
	"github.com/sirupsen/logrus"
)

// ErrInvalidHost describes an error for a hostname that is considered invalid if it's empty
var ErrInvalidHost = errors.New("Invalid hostname")

// ErrUnparsedURL describes an error that occours when the parsing process cannot be deemed as successful
var ErrUnparsedURL = errors.New("URL is not parsed correctly")

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

	newPath = filepath.Join(storeDir, dirPath)
	err = os.MkdirAll(newPath, os.ModePerm)
	return
}

// NewProcess creates a new transcoding process for ffmpeg
func NewProcess(URI string, spec *config.Specification) (*exec.Cmd, *Stream, string) {
	dirPath, newPath, err := createDirectoryForURI(URI, spec.StoreDir)
	if err != nil {
		logrus.Error("Error happened while getting directory name", dirPath)
		return nil, nil, ""
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
		"segment",
		"-segment_list_flags",
		"live",
		"-segment_time",
		"1",
		"-segment_list_size",
		"3",
		"-segment_format",
		"mpegts",
		"-segment_list",
		fmt.Sprintf("%s/index.m3u8", newPath),
		"-segment_list_type",
		"m3u8",
		"-segment_list_entry_prefix",
		fmt.Sprintf("/stream/%s/", dirPath),
		newPath+"/%d.ts",
	)
	stream := Stream{
		CMD:  cmd,
		Mux:  &sync.Mutex{},
		Path: fmt.Sprintf("/%s/index.m3u8", filepath.Join("stream", dirPath)),
		Streak: hotstreak.New(hotstreak.Config{
			Limit:      10,
			HotWait:    time.Minute * 2,
			ActiveWait: time.Minute * 4,
		}).Activate(),
		OriginalURI: URI,
	}
	return cmd, &stream, fmt.Sprintf("%s/index.m3u8", newPath)
}
