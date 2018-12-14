package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Roverr/hotstreak"
	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/Roverr/rtsp-stream/core/streaming"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// ErrNoStreamFn is used to create dynamic errors for unknown hosts requested as stream
var ErrNoStreamFn = func(path string) error { return fmt.Errorf("%s is not a known stream", path) }

// ErrStreamAlreadyActive is an error describing that we cannot restart the stream because it's already running
var ErrStreamAlreadyActive = errors.New("Stream is already active")

// streams keeps track of the streams based on their unique url combination
var streams = map[string]streaming.Stream{}

// streamDto describes an uri where the client can access the stream
type streamDto struct {
	URI string `json:"uri"`
}

// summariseDto describes each stream and their state of running
type summariseDto struct {
	Running bool   `json:"running"`
	URI     string `json:"uri"`
}

// restartStream is used when a stream is stopped but it gets a new request
func restartStream(spec *config.Specification, path string) error {
	stream, ok := streams[path]
	if !ok {
		return ErrNoStreamFn(path)
	}
	if stream.Streak.IsActive() {
		return ErrStreamAlreadyActive
	}
	stream.Mux.Lock()
	defer stream.Mux.Unlock()
	stream.CMD, _, _ = streaming.NewProcess(stream.OriginalURI, spec)
	stream.Streak.Activate()
	go func() {
		logrus.Infof("%s has been restarted", path)
		err := stream.CMD.Run()
		if err != nil {
			logrus.Error(err)
		}
	}()
	return nil
}

// getListStreamHandler returns the handler for the list endpoint
func getListStreamHandler() func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		dto := []*summariseDto{}
		for key, stream := range streams {
			dto = append(dto, &summariseDto{URI: fmt.Sprintf("/stream/%s/index.m3u8", key), Running: stream.Streak.IsActive()})
		}
		b, err := json.Marshal(dto)
		if err != nil {
			http.Error(w, "Internal server error", 500)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	}
}

// getStartStreamHandler returns an HTTP handler for the /start endpoint
func getStartStreamHandler(spec *config.Specification) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		// Parse request
		uri, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Invalid body", 400)
			return
		}
		var dto streamDto
		if err = json.Unmarshal(uri, &dto); err != nil {
			http.Error(w, "Invalid body", 400)
			return
		}
		if dto.URI == "" {
			http.Error(w, "Empty URI", 400)
			return
		}

		if _, err := url.Parse(dto.URI); err != nil {
			http.Error(w, "Invalid URI", 400)
			return
		}

		// Calculate directory from URI
		dir, err := streaming.GetURIDirectory(dto.URI)
		if err != nil {
			logrus.Error(err)
			http.Error(w, "Could not create directory for URI", 500)
			return
		}
		if s, ok := streams[dir]; ok {
			// If transcoding is not running, spin it back up
			if !s.Streak.IsActive() {
				err := restartStream(spec, dir)
				if err != nil {
					logrus.Error(err)
					http.Error(w, "Unexpected error", 500)
					return
				}
			}
			// If the stream is already running return its path
			b, err := json.Marshal(streamDto{URI: s.Path})
			if err != nil {
				http.Error(w, "Unexpected error", 500)
				return
			}
			w.Header().Add("Content-Type", "application/json")
			w.Write(b)
			return
		}
		streamRunning := make(chan bool)
		defer close(streamRunning)
		errorIssued := make(chan bool)
		defer close(errorIssued)
		go func() {
			logrus.Infof("%s started processing", dir)
			cmd, path, physicalPath := streaming.NewProcess(dto.URI, spec)
			streams[dir] = streaming.Stream{
				CMD:  cmd,
				Mux:  &sync.Mutex{},
				Path: fmt.Sprintf("/%s/index.m3u8", path),
				Streak: hotstreak.New(hotstreak.Config{
					Limit:      10,
					HotWait:    time.Minute * 2,
					ActiveWait: time.Minute * 4,
				}).Activate(),
				OriginalURI: dto.URI,
			}
			go func() {
				for {
					_, err := os.Stat(physicalPath)
					if err != nil {
						<-time.After(25 * time.Millisecond)
						continue
					}
					streamRunning <- true
					return
				}
			}()
			if err := cmd.Run(); err != nil {
				logrus.Error(err)
				errorIssued <- true
			}
		}()

		select {
		case <-time.After(time.Second * 10):
			http.Error(w, "Timeout error", 408)
			return
		case <-streamRunning:
			s := streams[dir]
			b, err := json.Marshal(streamDto{URI: s.Path})
			if err != nil {
				http.Error(w, "Unexpected error", 500)
				return
			}
			w.Header().Add("Content-Type", "application/json")
			w.Write(b)
		case <-errorIssued:
			http.Error(w, "Unexpected error", 500)
			return
		}
	}
}

// determinesHost is for parsing out the host from the storage path
func determineHost(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 1 {
		return parts[1]
	}
	return ""
}

// cleanUnusedProcesses is for stopping streams that are running despite having no viewers
func cleanUnusedProcesses() {
	for name, data := range streams {
		// If the streak is active, there is no need for stopping
		if data.Streak.IsActive() {
			logrus.Debugf("%s is active, skipping cleaning process", name)
			continue
		}
		logrus.Infof("%s is getting cleaned", name)
		data.Mux.Lock()
		defer data.Mux.Unlock()
		if err := data.CMD.Process.Kill(); err != nil {
			if strings.Contains(err.Error(), "process already finished") {
				logrus.Infof("\n%s is cleaned", name)
				continue
			}
			logrus.Error(err)
		}
		logrus.Infof("\n%s is cleaned", name)
	}
}

// GetRouter returns the return for the application
func GetRouter(config *config.Specification) *httprouter.Router {
	fileServer := http.FileServer(http.Dir(config.StoreDir))
	router := httprouter.New()
	if config.ListEndpoint {
		router.GET("/list", getListStreamHandler())
	}
	router.POST("/start", getStartStreamHandler(config))
	router.GET("/stream/*filepath", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		defer fileServer.ServeHTTP(w, req)
		filepath := ps.ByName("filepath")
		req.URL.Path = filepath
		hostKey := determineHost(filepath)
		s, ok := streams[hostKey]
		if !ok {
			return
		}
		if s.Streak.IsActive() {
			s.Streak.Hit()
			return
		}
		if err := restartStream(config, hostKey); err != nil {
			logrus.Error(err)
			return
		}
		s.Streak.Activate().Hit()
	})

	// Start cleaning process in the background
	go func() {
		for {
			<-time.After(config.CleanupTime)
			cleanUnusedProcesses()
		}
	}()

	return router
}
