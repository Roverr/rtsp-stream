package core

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Specification describes the application context settings
type Specification struct {
	Debug       bool          `default:"false" envconfig:"debug"`        // Indicates if debug log should be enabled or not
	Port        int           `envconfig:"port" default:"8080"`          // Port that the application listens on
	CleanupTime time.Duration `envconfig:"cleanup_time" default:"2m0s"`  // Time period between process cleaning
	StoreDir    string        `envconfig:"store_dir" default:"./videos"` // Directory to store / service video chunks
}

// InitConfig is to initalise the config
func InitConfig() *Specification {
	var s Specification
	err := envconfig.Process("RTSP_STREAM", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	return &s
}
