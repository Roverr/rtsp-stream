package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/Roverr/rtsp-stream/core/auth"
	"github.com/Roverr/rtsp-stream/core/blacklist"
	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/julienschmidt/httprouter"
	"github.com/riltech/streamer"
	"github.com/sirupsen/logrus"
)

// ErrUnexpected describes an unexpected error
var ErrUnexpected = errors.New("Unexpected error")

// ErrTimeout describes an error related to timing out
var ErrTimeout = errors.New("Timeout error")

// StreamDTO describes an uri where the client can access the stream
type StreamDTO struct {
	URI   string `json:"uri"`
	Alias string `json:"alias"`
}

// StopDTO describes a DTO for the /remove and /stop endpoints
type StopDTO struct {
	ID     string `json:"id"`
	Alias  string `json:"alias"`
	Wait   bool   `json:"wait"`
	Remove bool   `json:"remove"`
}

// SummariseDTO describes each stream and their state of running
type SummariseDTO struct {
	Running bool   `json:"running"`
	URI     string `json:"uri"`
	ID      string `json:"id"`
	Alias   string `json:"alias"`
}

// IController describes main functions for the controller
type IController interface {
	marshalValidatedURI(dto *StreamDTO, body io.Reader) error                             // marshals and validates request body for /start
	getIDByPath(path string) string                                                       // determines ID from the file access URL
	isAuthenticated(r *http.Request, endpoint string) bool                                // enforces JWT authentication if config is enabled
	stopInactiveStreams()                                                                 // used periodically to stop streams
	sendError(w http.ResponseWriter, err error, status int)                               // used by Handlers to send out errors
	sendStart(w http.ResponseWriter, success bool, stream *streamer.Stream, alias string) // used by start to send out response
	ListStreamHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params)        // handler - GET /list
	StartStreamHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params)       // handler - POST /start
	StaticFileHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params)        // handler - GET /stream/{id}/{file}
	StopStreamHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params)        // handler - POST /stop
	ExitPreHook() chan bool                                                               // runs before the application exits to clean up
}

// Controller holds all handler functions for the API
type Controller struct {
	spec       *config.Specification
	streams    map[string]*streamer.Stream
	index      map[string]string
	alias      map[string]string
	preload    map[string]string
	blacklist  blacklist.IList
	fileServer http.Handler
	timeout    time.Duration
	jwt        auth.JWT
}

// Type check
var _ IController = (*Controller)(nil)

// NewController creates a new instance of Controller
func NewController(spec *config.Specification, fileServer http.Handler) *Controller {
	provider, err := auth.NewJWTProvider(spec.Auth)
	if err != nil {
		logrus.Fatal("Could not create new JWT provider: ", err)
	}
	ctrl := &Controller{
		spec,
		map[string]*streamer.Stream{},
		map[string]string{},
		map[string]string{},
		map[string]string{},
		(*blacklist.List)(nil),
		fileServer,
		time.Second * 15,
		provider,
	}
	if spec.BlacklistEnabled {
		ctrl.blacklist = blacklist.NewList(spec.BlacklistTime, spec.BlacklistLimit)
	}
	if spec.CleanupEnabled {
		go func() {
			for {
				<-time.After(spec.CleanupTime)
				ctrl.stopInactiveStreams()
			}
		}()
	}

	// retain preloads
	for _, item := range spec.EndpointYML.Listen {
		if item.Enabled {
			ctrl.preload[item.Alias] = item.Uri
		}
	}

	return ctrl
}

// marshalValidateURI is for validiting that the URI is in a valid format
// and marshaling it into the dto pointer
func (c *Controller) marshalValidatedURI(dto *StreamDTO, body io.Reader) error {
	uri, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(uri, dto); err != nil {
		return err
	}

	if _, err := url.Parse(dto.URI); err != nil {
		return errors.New("Invalid URI")
	}
	return nil
}

// getIDByPath is for parsing out the unique ID of the stream from URL path
func (c *Controller) getIDByPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 1 {
		return parts[1]
	}
	return ""
}

// isAuthenticated is for checking if the user's request is valid or not
// from a given authentication strategy's perspective
func (c *Controller) isAuthenticated(r *http.Request, endpoint string) bool {
	if !c.spec.JWTEnabled {
		return true
	}
	token, claims := c.jwt.Validate(r.Header.Get("Authorization"))
	if token == nil || !token.Valid {
		return false
	}
	switch endpoint {
	case "list":
		if c.spec.Endpoints.List.Secret == "" {
			return true
		}
		return claims.Secret == c.spec.Endpoints.List.Secret
	case "start":
		if c.spec.Endpoints.Start.Secret == "" {
			return true
		}
		return claims.Secret == c.spec.Endpoints.Start.Secret
	case "stop":
		if c.spec.Endpoints.Stop.Secret == "" {
			return true
		}
		return claims.Secret == c.spec.Endpoints.Stop.Secret
	case "static":
		if c.spec.Endpoints.Static.Secret == "" {
			return true
		}
		return claims.Secret == c.spec.Endpoints.Static.Secret
	}
	return true
}

// stopInactiveStreams is for stopping all transcoding for streams that are not watched anymore
func (c *Controller) stopInactiveStreams() {
	for name, stream := range c.streams {
		// If the streak is active, there is no need for stopping
		if stream.Streak.IsActive() {
			logrus.Infof("%s is active. Skipping. | Inactivity cleaning", name)
			continue
		}
		if !stream.Running {
			logrus.Debugf("%s is not running. Skipping. | Inactivity cleaning", name)
			continue
		}
		logrus.Infof("%s is being stopped | Inactivity cleaning", name)
		if err := stream.Stop(); err != nil {
			logrus.Error(err)
		}
		logrus.Infof("%s is stopped | Inactivity cleaning", name)
	}
}

// sendError sends an error to the client
func (c *Controller) sendError(w http.ResponseWriter, err error, status int) {
	w.Header().Add("Content-Type", "application/json")
	b, _ := json.Marshal(struct {
		Error string `json:"error"`
	}{err.Error()})
	w.WriteHeader(status)
	w.Write(b)
}

// sendStart sends response for clients calling /start
func (c *Controller) sendStart(w http.ResponseWriter, success bool, stream *streamer.Stream, alias string) {
	if !stream.Running {
		logrus.Debugln("Sending out error for request timeout | StartHandler")
		c.sendError(w, ErrTimeout, http.StatusRequestTimeout)
		return
	}
	logrus.Infof("%s started processing | StartHandler", stream.OriginalURI)

	// use alias if provided
	name := stream.ID
	if len(alias) > 0 {
		name = alias
	}
	URI := fmt.Sprintf("/stream/%s/index.m3u8", name)

	b, _ := json.Marshal(SummariseDTO{URI: URI, Running: true, ID: stream.ID, Alias: alias})
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

// ListStreamHandler is the HTTP handler of the GET /list call
func (c *Controller) ListStreamHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !c.isAuthenticated(r, "list") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	dto := []*SummariseDTO{}

	// active streams
	for key, stream := range c.streams {
		aliasName := ""
		newKey := key
		for name, id := range c.alias {
			if id == stream.ID {
				aliasName = name
				newKey = name
			}
		}
		dto = append(dto, &SummariseDTO{
			URI:     fmt.Sprintf("/stream/%s/index.m3u8", newKey),
			Running: stream.Streak.IsActive(),
			ID:      stream.ID,
			Alias:   aliasName,
		})
	}

	// preload streams
	for name := range c.preload {
		dto = append(dto, &SummariseDTO{
			URI:     fmt.Sprintf("/stream/%s/index.m3u8", name),
			Running: false,
			ID:      "",
			Alias:   name,
		})
	}

	b, err := json.Marshal(dto)
	if err != nil {
		c.sendError(w, ErrUnexpected, http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

// StopStreamHandler is the HTTP handler of the stop stream request - POST /stop
func (c *Controller) StopStreamHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if !c.isAuthenticated(r, "stop") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dto := StopDTO{}
	err = json.Unmarshal(b, &dto)
	if err != nil {
		logrus.Error(err)
		c.sendError(w, err, http.StatusInternalServerError)
		return
	}

	if dto.ID == "" && len(dto.Alias) == 0 {
		c.sendError(w, err, http.StatusNotFound)
		return
	}

	if len(dto.Alias) > 0 {
		// redirect alias if used
		newid, ok := c.alias[dto.Alias]
		if ok {
			dto.ID = newid
		}
	}

	if s, ok := c.streams[dto.ID]; ok {
		logrus.Infof("%s is being stopped | StopStreamHandler", dto.ID)
		err := s.Stop()
		if err != nil {
			logrus.Error(err)
			c.sendError(w, err, http.StatusInternalServerError)
			return
		}
		if dto.Remove {
			delete(c.index, s.OriginalURI)
			delete(c.streams, dto.ID)
		}
	}
	logrus.Debugf("%s is stopped | StopStreamHandler", dto.ID)
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) startPreloadStream(Alias string, URI string) {
	logrus.Debugf("%s is being initialized", URI)

	_, knownStream := c.index[URI]
	if knownStream {
		return
	}

	stream, id := streamer.NewStream(
		URI,
		c.spec.StoreDir,
		c.spec.KeepFiles,
		c.spec.Audio,
		streamer.ProcessLoggingOpts{
			Enabled:    c.spec.ProcessLogging.Enabled,
			Compress:   c.spec.ProcessLogging.Compress,
			Directory:  c.spec.ProcessLogging.Directory,
			MaxAge:     c.spec.ProcessLogging.MaxAge,
			MaxBackups: c.spec.ProcessLogging.MaxBackups,
			MaxSize:    c.spec.ProcessLogging.MaxSize,
		},
		25*time.Second,
	)

	streamName := id
	stream.Start().Wait()
	if !stream.Running {
		if c.blacklist.AddOrIncrease(URI).IsBanned(URI) {
			delete(c.preload, Alias)
		}
		return
	}

	c.streams[id] = stream
	c.index[URI] = id
	if len(Alias) > 0 {
		c.alias[Alias] = id
		streamName = Alias
	}
	c.blacklist.Remove(URI)
	delete(c.preload, Alias)

	logrus.Infoln("started stream /stream/" + streamName + "/index.m3u8")
}

// StartStreamHandler is an HTTP handler for the POST /start endpoint
func (c *Controller) StartStreamHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if !c.isAuthenticated(r, "start") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	var dto StreamDTO
	if err := c.marshalValidatedURI(&dto, r.Body); err != nil {
		logrus.Error(err)
		c.sendError(w, err, http.StatusBadRequest)
		return
	}
	logrus.Debugf("%s is requested on /start | StartHandler", dto.URI)
	if c.blacklist.IsBanned(dto.URI) {
		logrus.Infof("%s is rejected because of blacklist | StartHandler", dto.URI)
		c.sendError(w, fmt.Errorf("%s cannot be started", dto.URI), http.StatusTooManyRequests)
		return
	}
	index, knownStream := c.index[dto.URI]
	if knownStream {
		stream, ok := c.streams[index]
		if !ok {
			logrus.Errorf("Missing index for URI: %s | StartHandler", dto.URI)
			c.sendError(w, ErrUnexpected, http.StatusInternalServerError)
			return
		}

		// look for an alias since its known
		alias := ""
		for key, item := range c.alias {
			if item == stream.ID {
				alias = key
			}
		}

		if stream.Running {
			c.sendStart(w, true, stream, alias)
			return
		}
		stream.Restart().Wait()
		c.sendStart(w, stream.Running, stream, alias)
		return
	}

	stream, id := streamer.NewStream(
		dto.URI,
		c.spec.StoreDir,
		c.spec.KeepFiles,
		c.spec.Audio,
		streamer.ProcessLoggingOpts{
			Enabled:    c.spec.ProcessLogging.Enabled,
			Compress:   c.spec.ProcessLogging.Compress,
			Directory:  c.spec.ProcessLogging.Directory,
			MaxAge:     c.spec.ProcessLogging.MaxAge,
			MaxBackups: c.spec.ProcessLogging.MaxBackups,
			MaxSize:    c.spec.ProcessLogging.MaxSize,
		},
		25*time.Second,
	)
	stream.Start().Wait()
	if stream.Running {
		c.streams[id] = stream
		c.index[dto.URI] = id
		c.blacklist.Remove(dto.URI)
		// only add an alias if provided
		if len(dto.Alias) > 0 {
			c.alias[dto.Alias] = id
		}
	} else {
		c.blacklist.AddOrIncrease(dto.URI)
	}
	c.sendStart(w, stream.Running, stream, dto.Alias)
}

func (c *Controller) shouldRedirectAlias(alias string, filepath string) (string, bool) {
	id, ok := c.alias[alias]
	if !ok {
		return "", false
	}

	url := strings.Join(
		[]string{
			"/stream",
			id,
			path.Base(filepath),
		}, "/")

	return url, true
}

// StaticFileHandler is HTTP handler for direct file requests
func (c *Controller) StaticFileHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	if !c.isAuthenticated(req, "static") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	defer c.fileServer.ServeHTTP(w, req)
	filepath := ps.ByName("filepath")
	req.URL.Path = filepath
	id := c.getIDByPath(filepath)

	// start preload if registered
	uri, ok := c.preload[id]
	if ok {
		logrus.Infoln("starting preload " + id + " now")
		c.startPreloadStream(id, uri)
	}

	// redirect alias if used

	if url, ok := c.shouldRedirectAlias(id, filepath); ok {
		logrus.Infoln("redirecting alias " + id + " to " + url)
		http.Redirect(w, req, url, 302)
		return
	}

	stream, ok := c.streams[id]
	if !ok {
		return
	}
	if stream.Streak.IsActive() || stream.Running {
		stream.Streak.Hit()
		return
	}
	logrus.Debugf("%s is getting restarted via file requests | FileHandler", id)
	stream.Restart().Wait()
}

// ExitPreHook is a function that can recognise when the application is being closed
// and cleans up all background running processes
func (c *Controller) ExitPreHook() chan bool {
	done := make(chan bool)
	ch := make(chan os.Signal, 3)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ch
		for uri, strm := range c.streams {
			logrus.Debugf("Closing processing of %s", uri)
			if err := strm.Stop(); err != nil {
				logrus.Error(err)
				return
			}
			logrus.Debugf("Succesfully closed processing for %s", uri)
		}
		done <- true
	}()
	return done
}
