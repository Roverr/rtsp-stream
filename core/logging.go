package core

import (
	"os"

	"github.com/sirupsen/logrus"
)

// SetupLogger sets the logger for the proper settings based on the environment
func SetupLogger(spec *Specification) {
	logrus.SetOutput(os.Stdout)
	if spec.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		return
	}
	logrus.SetLevel(logrus.InfoLevel)
}
