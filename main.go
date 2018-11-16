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
	"github.com/julienschmidt/httprouter"
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
			fmt.Printf("\n%s is active, skipping cleaning process\n", name)
			continue
		}
		fmt.Printf("\n%s is getting cleaned\n", name)
		data.Mux.Lock()
		defer data.Mux.Unlock()
		if err := data.CMD.Process.Kill(); err != nil {
			if strings.Contains(err.Error(), "process already finished") {
				fmt.Printf("\n%s is cleaned", name)
				continue
			}
			log.Fatal(err)
		}
		fmt.Printf("\n%s is cleaned", name)
	}
}

func startHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
	dir, err := getURIDirectory(dto.URI)
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
		fmt.Println("Starting processing of", dir)
		cmd, path := newProcess(dto.URI)
		streams[dir] = stream{
			CMD:  cmd,
			Mux:  &sync.Mutex{},
			Path: fmt.Sprintf("/%s/index.m3u8", path),
			streak: hotstreak.New(hotstreak.Config{
				Limit:      10,
				HotWait:    time.Minute * 1,
				ActiveWait: time.Minute * 2,
			}).Activate(),
		}
		streamRunning <- true
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
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

func cleanUpHandler() chan bool {
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

func cleanup() {
	for uri, strm := range streams {
		fmt.Printf("\nClosing processing of %s\n", uri)
		strm.Mux.Lock()
		strm.streak.Deactivate()
		defer strm.Mux.Unlock()
		if err := strm.CMD.Process.Kill(); err != nil {
			if strings.Contains(err.Error(), "process already finished") {
				continue
			}
			log.Fatal(err)
		}
		fmt.Println("Succesfully closed processing for", uri)
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
	done := cleanUpHandler()
	router := httprouter.New()
	router.POST("/start", startHandler)

	fileServer := http.FileServer(http.Dir("./videos"))

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
			<-time.After(time.Minute * 2)
			cleanUnusedProcesses()
		}
	}()
	log.Fatal(http.ListenAndServe(":8080", router))
	<-done
}
