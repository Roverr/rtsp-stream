package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Roverr/hotstreak"
	"github.com/Roverr/rtsp-stream/core"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

var streams = map[string]stream{}

type stream struct {
	CMD    *exec.Cmd
	Mux    *sync.Mutex
	Path   string
	streak *hotstreak.Hotstreak
}

type streamDto struct {
	URI string `json:"uri"`
}

func cleanUnusedProcesses() {
	for name, data := range streams {
		if data.streak.IsActive() {
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
func getStartStreamHandler(spec *core.Specification) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		dir, err := core.GetURIDirectory(dto.URI)
		if err != nil {
			http.Error(w, "Could not create directory for URI", 500)
			return
		}
		if s, ok := streams[dir]; ok {
			b, err := json.Marshal(streamDto{URI: s.Path})
			if err != nil {
				http.Error(w, "Unexpected error", 500)
				return
			}
			w.Write(b)
			return
		}
		streamRunning := make(chan bool)
		defer close(streamRunning)
		go func() {
			logrus.Infof("Starting processing of %s", dir)
			cmd, path := core.NewProcess(dto.URI, spec)
			streams[dir] = stream{
				CMD:  cmd,
				Mux:  &sync.Mutex{},
				Path: fmt.Sprintf("/%s/index.m3u8", path),
				streak: hotstreak.New(hotstreak.Config{
					Limit:      10,
					HotWait:    time.Minute * 2,
					ActiveWait: time.Minute * 4,
				}).Activate(),
			}
			streamRunning <- true
			if err := cmd.Run(); err != nil {
				logrus.Error(err)
			}
		}()
		<-streamRunning
		s := streams[dir]
		b, err := json.Marshal(streamDto{URI: s.Path})
		if err != nil {
			http.Error(w, "Unexpected error", 500)
			return
		}
		w.Write(b)
	}
}

func exitHandler() chan bool {
	done := make(chan bool)
	c := make(chan os.Signal, 3)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
		done <- true
	}()
	return done
}

func cleanProcess(strm stream) error {
	strm.Mux.Lock()
	strm.streak.Deactivate()
	defer strm.Mux.Unlock()
	if err := strm.CMD.Process.Kill(); err != nil {
		if strings.Contains(err.Error(), "process already finished") {
			return nil
		}
		return err
	}
	return nil
}
func cleanup() {
	for uri, strm := range streams {
		logrus.Debugf("Closing processing of %s", uri)
		if err := cleanProcess(strm); err != nil {
			logrus.Debugf("Could not close %s", uri)
			logrus.Error(err)
			return
		}
		logrus.Debugf("Succesfully closed processing for %s", uri)
	}
}

func determineHost(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 1 {
		return parts[1]
	}
	return ""
}

func main() {
	config := core.InitConfig()
	core.SetupLogger(config)
	done := exitHandler()

	router := httprouter.New()
	router.POST("/start", getStartStreamHandler(config))

	fileServer := http.FileServer(http.Dir(config.StoreDir))

	router.GET("/stream/*filepath", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		filepath := ps.ByName("filepath")
		req.URL.Path = filepath
		fileServer.ServeHTTP(w, req)
		if s, ok := streams[determineHost(filepath)]; ok {
			s.streak.Hit()
		}
	})

	go func() {
		for {
			<-time.After(config.CleanupTime)
			cleanUnusedProcesses()
		}
	}()
	logrus.Infof("RTSP-STREAM started on %d", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), router))
	<-done
}
