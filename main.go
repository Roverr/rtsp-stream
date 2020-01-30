package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"

	"github.com/Roverr/rtsp-stream/core"
	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/sirupsen/logrus"
)

func main() {
	config := config.InitConfig()
	core.SetupLogger(config)
	fileServer := http.FileServer(http.Dir(config.StoreDir))
	router := httprouter.New()
	controllers := core.NewController(config, fileServer)
	if config.ListEndpoint {
		router.GET("/list", controllers.ListStreamHandler)
	}
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})
	router.POST("/start", controllers.StartStreamHandler)
	router.GET("/stream/*filepath", controllers.StaticFileHandler)
	router.POST("/stop", controllers.StopStreamHandler)
	done := controllers.ExitPreHook()
	handler := cors.AllowAll().Handler(router)
	if config.CORS.Enabled {
		handler = cors.New(cors.Options{
			AllowCredentials: config.CORS.AllowCredentials,
			AllowedOrigins:   config.CORS.AllowedOrigins,
			MaxAge:           config.CORS.MaxAge,
		}).Handler(router)
	}
	logrus.Infof("rtsp-stream transcoder started on %d | MainProcess", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), handler))
	<-done
	os.Exit(0)
}
