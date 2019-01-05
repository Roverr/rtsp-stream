package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Roverr/hotstreak"

	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/Roverr/rtsp-stream/core/streaming"
	"github.com/brianvoe/gofakeit"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

type generatedStream struct {
	strm    streaming.Stream
	dirPath string
}

func generateURI() string {
	return fmt.Sprintf("rtps://%s:%s@192.168.0.1/%s/Channels/001", gofakeit.Word(), gofakeit.Word(), gofakeit.Word())
}

func generateStream(hs *hotstreak.Hotstreak) generatedStream {
	uri := generateURI()
	dirPath, _ := streaming.GetURIDirectory(uri)
	streak := hs
	if hs == nil {
		streak = hotstreak.New(hotstreak.Config{
			Limit:      10,
			HotWait:    time.Minute * 3,
			ActiveWait: time.Minute * 4,
		})
	}
	return generatedStream{
		strm: streaming.Stream{
			Mux:         &sync.Mutex{},
			Path:        fmt.Sprintf("/stream/%s/index.m3u8", dirPath),
			OriginalURI: uri,
			Streak:      streak,
		},
		dirPath: dirPath,
	}
}
func TestController(t *testing.T) {
	cfg := config.InitConfig()
	fileServer := http.FileServer(http.Dir(cfg.StoreDir))
	ctrls := NewController(cfg, fileServer)
	router := httprouter.New()
	router.GET("/list", ctrls.ListStreamHandler)
	router.POST("/start", ctrls.StartStreamHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Should get empty list if no streams available", func(t *testing.T) {
		res, err := http.Get(fmt.Sprintf("%s/list", server.URL))
		assert.Nil(t, err)
		b, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		var result []streamDto
		assert.Nil(t, json.Unmarshal(b, &result))
		assert.Empty(t, result)
	})

	t.Run("Should get streams back if they are available", func(t *testing.T) {
		generated := generateStream(nil)
		ctrls.streams = map[string]*streaming.Stream{
			generated.dirPath: &generated.strm,
		}
		res, err := http.Get(fmt.Sprintf("%s/list", server.URL))
		assert.Nil(t, err)
		b, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		var result []streamDto
		assert.Nil(t, json.Unmarshal(b, &result))
		assert.NotEmpty(t, result)
		assert.Equal(t, result[0].URI, generated.strm.Path)
	})

	t.Run("Should be able to get back already running streams instantly", func(t *testing.T) {
		generated := generateStream(nil)
		generated.strm.Streak.Activate()
		ctrls.streams = map[string]*streaming.Stream{
			generated.dirPath: &generated.strm,
		}
		generated.strm.Streak.Hit()
		dto := streamDto{
			URI: generated.strm.OriginalURI,
		}
		b, err := json.Marshal(dto)
		assert.Nil(t, err)
		res, err := http.Post(fmt.Sprintf("%s/start", server.URL), "application/json", bytes.NewBuffer(b))
		assert.Nil(t, err)
		b, err = ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		var result streamDto
		fmt.Println(string(b))
		assert.Nil(t, json.Unmarshal(b, &result))
		assert.Equal(t, result.URI, generated.strm.Path)
	})
}
