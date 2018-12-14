package config

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// CORS is the ptions for cross origin handling
type CORS struct {
	Enabled          bool     `envconfig:"CORS_ENABLED" default:"false"`           // Indicates if cors should be handled as configured or as default
	AllowedOrigins   []string `envconfig:"CORS_ALLOWED_ORIGINS" default:""`        // A list of origins a cross-domain request can be executed from.
	AllowCredentials bool     `envconfig:"CORS_ALLOW_CREDENTIALS" default:"false"` // Indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates.
	MaxAge           int      `envconfig:"CORS_MAX_AGE" default:"0"`               // Indicates how long (in seconds) the results of a preflight request can be cached.
}

// Specification describes the application context settings
type Specification struct {
	Debug        bool          `envconfig:"DEBUG" default:"false"`         // Indicates if debug log should be enabled or not
	Port         int           `envconfig:"PORT" default:"8080"`           // Port that the application listens on
	CleanupTime  time.Duration `envconfig:"CLEANUP_TIME" default:"2m0s"`   // Time period between process cleaning
	StoreDir     string        `envconfig:"STORE_DIR" default:"./videos"`  // Directory to store / service video chunks
	ListEndpoint bool          `envconfig:"LIST_ENDPOINT" default:"false"` // Turns on / off the stream listing endpoint feature

	CORS
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
