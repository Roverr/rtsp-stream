package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/Roverr/hotstreak"
	"github.com/julienschmidt/httprouter"

	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/Roverr/rtsp-stream/core/streaming"
	"github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())
	os.Exit(m.Run())
}

type generatedStream struct {
	strm    streaming.Stream
	dirPath string
}

func generateURI() string {
	return fmt.Sprintf("rtps://%s:%s@192.168.0.1/%s/Channels/001", gofakeit.Word(), gofakeit.Word(), gofakeit.Word())
}

func generateStream(hs *hotstreak.Hotstreak, URI string) generatedStream {
	uri := URI
	if URI == "" {
		uri = generateURI()
	}
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

type mockManager struct {
	resolve bool
	instead *func(physicalPath string) chan bool
}

var _ IManager = (*mockManager)(nil)

func (m mockManager) Start(cmd *exec.Cmd, physicalPath string) chan bool {
	if m.instead != nil {
		return (*m.instead)(physicalPath)
	}
	streamResolved := make(chan bool, 1)
	streamResolved <- m.resolve
	return streamResolved
}

type mockProcessor struct{}

var _ streaming.IProcessor = (*mockProcessor)(nil)

func (m mockProcessor) NewProcess(URI string) *exec.Cmd {
	return nil
}

func (m mockProcessor) NewStream(URI string) (*streaming.Stream, string) {
	generated := generateStream(nil, URI)
	return &generated.strm, generated.strm.Path
}

func (m mockProcessor) Restart(stream *streaming.Stream, path string) error {
	stream.Streak.Activate().Hit()
	return nil
}

func TestController(t *testing.T) {
	cfg := config.InitConfig()
	fileServer := http.FileServer(http.Dir(cfg.StoreDir))

	t.Run("Should get empty list if no streams available", func(t *testing.T) {
		ctrls := NewController(cfg, fileServer)
		router := httprouter.New()
		router.GET("/list", ctrls.ListStreamHandler)
		router.POST("/start", ctrls.StartStreamHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		res, err := http.Get(fmt.Sprintf("%s/list", server.URL))
		assert.Nil(t, err)
		b, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		var result []streamDto
		assert.Nil(t, json.Unmarshal(b, &result))
		assert.Empty(t, result)
	})

	t.Run("Should get streams back if they are available", func(t *testing.T) {
		ctrls := NewController(cfg, fileServer)
		router := httprouter.New()
		router.GET("/list", ctrls.ListStreamHandler)
		router.POST("/start", ctrls.StartStreamHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		generated := generateStream(nil, "")
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
		ctrls := NewController(cfg, fileServer)
		router := httprouter.New()
		router.GET("/list", ctrls.ListStreamHandler)
		router.POST("/start", ctrls.StartStreamHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		generated := generateStream(nil, "")
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
		assert.Nil(t, json.Unmarshal(b, &result))
		assert.Equal(t, result.URI, generated.strm.Path)
	})

	t.Run("Should be able to start stream correctly", func(t *testing.T) {
		ctrls := NewController(cfg, fileServer)
		ctrls.manager = mockManager{resolve: true}
		ctrls.processor = mockProcessor{}
		ctrls.streams = map[string]*streaming.Stream{}
		router := httprouter.New()
		router.GET("/list", ctrls.ListStreamHandler)
		router.POST("/start", ctrls.StartStreamHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		dto := streamDto{
			URI: generateURI(),
		}
		dir, err := streaming.GetURIDirectory(dto.URI)
		assert.Nil(t, err)
		b, err := json.Marshal(dto)
		assert.Nil(t, err)
		res, err := http.Post(fmt.Sprintf("%s/start", server.URL), "application/json", bytes.NewBuffer(b))
		assert.Nil(t, err)
		b, err = ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		var result streamDto
		assert.Nil(t, json.Unmarshal(b, &result))
		strm, ok := ctrls.streams[dir]
		assert.True(t, ok)
		assert.Equal(t, result.URI, strm.Path)
	})

	t.Run("Should be able to receive unexpected error if something happens", func(t *testing.T) {
		ctrls := NewController(cfg, fileServer)
		ctrls.manager = mockManager{resolve: false}
		ctrls.processor = mockProcessor{}
		ctrls.streams = map[string]*streaming.Stream{}
		router := httprouter.New()
		router.GET("/list", ctrls.ListStreamHandler)
		router.POST("/start", ctrls.StartStreamHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		dto := streamDto{
			URI: generateURI(),
		}
		b, err := json.Marshal(dto)
		assert.Nil(t, err)
		res, err := http.Post(fmt.Sprintf("%s/start", server.URL), "application/json", bytes.NewBuffer(b))
		assert.Nil(t, err)
		b, err = ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		var errDto ErrDTO
		assert.Nil(t, json.Unmarshal(b, &errDto))
		assert.Equal(t, ErrDTO{ErrUnexpected.Error()}, errDto)
	})

	t.Run("Should be able to timeout if the process takes too long", func(t *testing.T) {
		instead := func(physicalPath string) chan bool {
			return make(chan bool)
		}
		ctrls := NewController(cfg, fileServer)
		ctrls.timeout = time.Millisecond * 300
		ctrls.manager = mockManager{instead: &instead}
		ctrls.processor = mockProcessor{}
		ctrls.streams = map[string]*streaming.Stream{}
		router := httprouter.New()
		router.GET("/list", ctrls.ListStreamHandler)
		router.POST("/start", ctrls.StartStreamHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		dto := streamDto{
			URI: generateURI(),
		}
		b, err := json.Marshal(dto)
		assert.Nil(t, err)
		res, err := http.Post(fmt.Sprintf("%s/start", server.URL), "application/json", bytes.NewBuffer(b))
		assert.Nil(t, err)
		b, err = ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		var errDto ErrDTO
		assert.Nil(t, json.Unmarshal(b, &errDto))
		assert.Equal(t, ErrDTO{ErrTimeout.Error()}, errDto)
	})
}
