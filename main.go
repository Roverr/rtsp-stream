package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"

	"github.com/Roverr/rtsp-stream/core"
	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/sirupsen/logrus"
)

func main() {
	config := config.InitConfig()
	core.SetupLogger(config)
	router, ctrls := core.GetRouter(config)
	done := ctrls.ExitPreHook()
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
