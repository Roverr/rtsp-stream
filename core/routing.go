package core

import (
	"net/http"
	"strings"

	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/julienschmidt/httprouter"
)

// getIDByPath is for parsing out the unique ID of the stream from URL path
func getIDByPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 1 {
		return parts[1]
	}
	return ""
}

// GetRouter returns the return for the application
func GetRouter(config *config.Specification) (*httprouter.Router, *Controller) {
	fileServer := http.FileServer(http.Dir(config.StoreDir))
	router := httprouter.New()
	controllers := NewController(config, fileServer)
	if config.ListEndpoint {
		router.GET("/list", controllers.ListStreamHandler)
	}
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})
	router.POST("/start", controllers.StartStreamHandler)
	router.GET("/stream/*filepath", controllers.StaticFileHandler)
	return router, controllers
}
