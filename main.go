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
	"sync"
	"syscall"

	"github.com/julienschmidt/httprouter"
)

var streams = map[string]stream{}

type stream struct {
	CMD     *exec.Cmd
	Running bool
	Mux     *sync.Mutex
	Path    string
}
type streamDto struct {
	URI string `json:"uri"`
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	uri, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	var dto streamDto
	if err = json.Unmarshal(uri, &dto); err != nil {
		log.Fatal(err)
	}
	if dto.URI == "" {
		log.Fatal("Empty uri")
	}
	if _, ok := streams[dto.URI]; ok {
		return
	}
	streamRunning := make(chan bool)
	defer close(streamRunning)
	go func() {
		fmt.Println("Starting processing of ", dto.URI)
		cmd, path := newProcess(dto.URI)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		streams[dto.URI] = stream{
			CMD:     cmd,
			Running: true,
			Mux:     &sync.Mutex{},
			Path:    path,
		}
		streamRunning <- true
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	<-streamRunning
	s := streams[dto.URI]
	b, err := json.Marshal(streamDto{URI: s.Path})
	if err != nil {
		log.Fatal(err)
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
		strm.Running = false
		if err := strm.CMD.Process.Kill(); err != nil {
			log.Fatal(err)
		}
		strm.Mux.Unlock()
		fmt.Println("Succesfully closed processing for", uri)
	}
}

func main() {
	done := cleanUpHandler()
	router := httprouter.New()
	router.POST("/start", Index)
	router.ServeFiles("/stream/*filepath", http.Dir("./videos"))
	log.Fatal(http.ListenAndServe(":8080", router))
	<-done
}
