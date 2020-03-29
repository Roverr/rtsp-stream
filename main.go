package main

import (
	"context"
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
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})
	if config.EndpointYML.Endpoints.List.Enabled {
		router.GET("/list", controllers.ListStreamHandler)
		logrus.Infoln("list endpoint enabled | MainProcess")
	}
	if config.EndpointYML.Endpoints.Start.Enabled {
		router.POST("/start", controllers.StartStreamHandler)
		logrus.Infoln("start endpoint enabled | MainProcess")
	}
	if config.EndpointYML.Endpoints.Static.Enabled {
		router.GET("/stream/*filepath", controllers.StaticFileHandler)
		logrus.Infoln("static endpoint enabled | MainProcess")
	}
	if config.EndpointYML.Endpoints.Stop.Enabled {
		router.POST("/stop", controllers.StopStreamHandler)
		logrus.Infoln("stop endpoint enabled | MainProcess")
	}

	done := controllers.ExitPreHook()
	handler := cors.AllowAll().Handler(router)
	if config.CORS.Enabled {
		handler = cors.New(cors.Options{
			AllowCredentials: config.CORS.AllowCredentials,
			AllowedOrigins:   config.CORS.AllowedOrigins,
			MaxAge:           config.CORS.MaxAge,
		}).Handler(router)
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: handler,
	}
	go func() {
		logrus.Infof("rtsp-stream transcoder started on %d | MainProcess", config.Port)
		log.Fatal(srv.ListenAndServe())
	}()
	<-done
	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("HTTP server Shutdown: %v", err)
	}
	os.Exit(0)
}
