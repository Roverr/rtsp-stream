package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rs/cors"

	"github.com/Roverr/rtsp-stream/core"
	"github.com/sirupsen/logrus"
)

func main() {
	config := core.InitConfig()
	core.SetupLogger(config)
	done := core.ExitHandler()
	router := core.GetRouter(config)
	handler := cors.AllowAll().Handler(router)
	if config.CORS.Enabled {
		handler = cors.New(cors.Options{
			AllowCredentials: config.CORS.AllowCredentials,
			AllowedOrigins:   config.CORS.AllowedOrigins,
			MaxAge:           config.CORS.MaxAge,
		}).Handler(router)
	}
	logrus.Infof("RTSP-STREAM started on %d", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), handler))
	<-done
}
