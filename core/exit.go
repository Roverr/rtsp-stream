package core

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Roverr/rtsp-stream/core/streaming"
	"github.com/sirupsen/logrus"
)

// ExitHandler is a function that can recognise when the application is being closed
// and cleans up all background running processes
func ExitHandler() chan bool {
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

func cleanProcess(strm streaming.Stream) error {
	strm.Mux.Lock()
	strm.Streak.Deactivate()
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
