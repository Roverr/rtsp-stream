package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/Roverr/rtsp-stream/core/streaming"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// Controller holds all handler functions for the API
type Controller struct {
	spec       *config.Specification
	streams    map[string]*streaming.Stream
	fileServer http.Handler
}

// NewController creates a new instance of Controller
func NewController(spec *config.Specification, fileServer http.Handler) *Controller {
	return &Controller{spec, map[string]*streaming.Stream{}, fileServer}
}

// ListStreamHandler is the HTTP handler of the /list call
func (c *Controller) ListStreamHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	dto := []*summariseDto{}
	for key, stream := range c.streams {
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

// StartStreamHandler is an HTTP handler for the /start endpoint
func (c *Controller) StartStreamHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var dto streamDto
	if err := validateURI(&dto, r.Body); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// Calculate directory from URI
	dir, err := streaming.GetURIDirectory(dto.URI)
	if err != nil {
		logrus.Error(err)
		http.Error(w, "Could not create directory for URI", 500)
		return
	}
	if stream, ok := c.streams[dir]; ok {
		handleAlreadyRunningStream(w, stream, c.spec, dir)
		return
	}
	streamResolved := make(chan bool)
	defer close(streamResolved)
	go c.startStream(dto.URI, dir, c.spec, streamResolved)
	select {
	case <-time.After(time.Second * 10):
		http.Error(w, "Timeout error", 408)
		return
	case success := <-streamResolved:
		if !success {
			http.Error(w, "Unexpected error", 500)
			return
		}
		s := c.streams[dir]
		b, err := json.Marshal(streamDto{URI: s.Path})
		if err != nil {
			http.Error(w, "Unexpected error", 500)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	}
}

// ExitHandler is a function that can recognise when the application is being closed
// and cleans up all background running processes
func (c *Controller) ExitHandler() chan bool {
	done := make(chan bool)
	ch := make(chan os.Signal, 3)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ch
		c.cleanUp()
		os.Exit(0)
		done <- true
	}()
	return done
}

// cleanUp stops all running processes
func (c *Controller) cleanUp() {
	for uri, strm := range c.streams {
		logrus.Debugf("Closing processing of %s", uri)
		if err := strm.CleanProcess(); err != nil {
			logrus.Debugf("Could not close %s", uri)
			logrus.Error(err)
			return
		}
		logrus.Debugf("Succesfully closed processing for %s", uri)
	}
}

// cleanUnused is for stopping all transcoding for streams that are not watched anymore
func (c *Controller) cleanUnused() {
	for name, data := range c.streams {
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

// FileHandler is HTTP handler for direct file requests
func (c *Controller) FileHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	defer c.fileServer.ServeHTTP(w, req)
	filepath := ps.ByName("filepath")
	req.URL.Path = filepath
	hostKey := determineHost(filepath)
	s, ok := c.streams[hostKey]
	if !ok {
		return
	}
	if s.Streak.IsActive() {
		s.Streak.Hit()
		return
	}
	if err := s.Restart(c.spec, hostKey); err != nil {
		logrus.Error(err)
		return
	}
	s.Streak.Activate().Hit()
}

func (c *Controller) startStream(uri, dir string, spec *config.Specification, streamResolved chan<- bool) {
	logrus.Infof("%s started processing", dir)
	cmd, stream, physicalPath := streaming.NewProcess(uri, spec)
	c.streams[dir] = stream
	var once sync.Once
	go func() {
		for {
			_, err := os.Stat(physicalPath)
			if err != nil {
				<-time.After(25 * time.Millisecond)
				continue
			}
			once.Do(func() { streamResolved <- true })
			return
		}
	}()
	if err := cmd.Run(); err != nil {
		logrus.Error(err)
		once.Do(func() { streamResolved <- false })
	}
}
