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

// Auth describes information regarding authentication
type Auth struct {
	JWTEnabled    bool   `envconfig:"AUTH_JWT_ENABLED" default:"false"`      // Indicates if JWT authentication is enabled or not
	JWTSecret     string `envconfig:"AUTH_JWT_SECRET" default:"macilaci"`    // Secret of the JWT encryption
	JWTMethod     string `envconfig:"AUTH_JWT_METHOD" default:"secret"`      // Can be "secret" or "rsa", defines the decoding method
	JWTPubKeyPath string `envconfig:"AUTH_JWT_PUB_PATH" default:"./key.pub"` // Path to the public RSA key
}

// ProcessLogging describes information about the logging mechanism of the transcoding FFMPEG process
type ProcessLogging struct {
	Enabled    bool   `envconfig:"PROCESS_LOGGING" default:"true"`                     // Option to set logging for transcoding processes
	Directory  string `envconfig:"PROCESS_LOGGING_DIR" default:"/var/log/rtsp-stream"` // Directory for the logs
	MaxSize    int    `envconfig:"PROCESS_LOGGING_MAX_SIZE" default:"500"`             // Maximum size of kept logging files in megabytes
	MaxBackups int    `envconfig:"PROCESS_LOGGING_MAX_BACKUPS" default:"3"`            // Maximum number of old log files to retain
	MaxAge     int    `envconfig:"PROCESS_LOGGING_MAX_AGE" default:"7"`                // Maximum number of days to retain an old log file.
	Compress   bool   `envconfig:"PROCESS_LOGGING_COMPRESS" default:"true"`            // Indicates if the log rotation should compress the log files
}

// Process describes information regarding the transcoding process
type Process struct {
	CleanupTime time.Duration `envconfig:"CLEANUP_TIME" default:"2m0s"`  // Time period between process cleaning
	StoreDir    string        `envconfig:"STORE_DIR" default:"./videos"` // Directory to store / service video chunks
	KeepFiles   bool          `envconfig:"KEEP_FILES" default:"false"`   // Option for not deleting files
}

// Specification describes the application context settings
type Specification struct {
	Debug        bool `envconfig:"DEBUG" default:"false"`         // Indicates if debug log should be enabled or not
	Port         int  `envconfig:"PORT" default:"8080"`           // Port that the application listens on
	ListEndpoint bool `envconfig:"LIST_ENDPOINT" default:"false"` // Turns on / off the stream listing endpoint feature

	CORS
	Auth
	Process
	ProcessLogging
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
