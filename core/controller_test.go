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
		var result []StreamDto
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
		var result []StreamDto
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
		dto := StreamDto{
			URI: generated.strm.OriginalURI,
		}
		b, err := json.Marshal(dto)
		assert.Nil(t, err)
		res, err := http.Post(fmt.Sprintf("%s/start", server.URL), "application/json", bytes.NewBuffer(b))
		assert.Nil(t, err)
		b, err = ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		var result StreamDto
		assert.Nil(t, json.Unmarshal(b, &result))
		assert.Equal(t, result.URI, generated.strm.Path)
	})

	t.Run("Should be able to start stream correctly", func(t *testing.T) {
		ctrls := NewController(cfg, fileServer)
		ctrls.manager = mockManager{resolve: true}
		ctrls.processor = mockProcessor{}
		ctrls.streams = map[string]*streaming.Stream{}
		router := httprouter.New()
		router.POST("/start", ctrls.StartStreamHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		dto := StreamDto{
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
		var result StreamDto
		assert.Nil(t, json.Unmarshal(b, &result))
		strm, ok := ctrls.streams[dir]
		assert.True(t, ok)
		assert.Equal(t, result.URI, strm.Path)
	})

	t.Run("Should be able to restart stream correctly", func(t *testing.T) {
		ctrls := NewController(cfg, fileServer)
		ctrls.manager = mockManager{resolve: true}
		ctrls.processor = mockProcessor{}
		ctrls.streams = map[string]*streaming.Stream{}
		router := httprouter.New()
		router.POST("/start", ctrls.StartStreamHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		generated := generateStream(nil, "")
		generated.strm.Streak.Deactivate()
		ctrls.streams = map[string]*streaming.Stream{
			generated.dirPath: &generated.strm,
		}
		dto := StreamDto{
			URI: generated.strm.OriginalURI,
		}
		dir, err := streaming.GetURIDirectory(dto.URI)
		assert.Nil(t, err)
		b, err := json.Marshal(dto)
		assert.Nil(t, err)
		res, err := http.Post(fmt.Sprintf("%s/start", server.URL), "application/json", bytes.NewBuffer(b))
		assert.Nil(t, err)
		b, err = ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		var result StreamDto
		assert.Nil(t, json.Unmarshal(b, &result))
		strm, ok := ctrls.streams[dir]
		assert.True(t, ok)
		assert.Equal(t, result.URI, strm.Path)

		assert.True(t, ctrls.streams[generated.dirPath].Streak.IsActive())
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

		dto := StreamDto{
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

		dto := StreamDto{
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

	t.Run("Should be able to clean unusued streams", func(t *testing.T) {
		ctrls := NewController(cfg, fileServer)
		wg := &sync.WaitGroup{}
		wg.Add(1)

		generated := generateStream(nil, "")
		generated.strm.Streak.Deactivate()
		generated.strm.CMD = exec.Command("tail", "-f", "/dev/null")
		go func() {
			generated.strm.CMD.Run()
			wg.Done()
		}()
		activeGenerated := generateStream(nil, "")
		activeGenerated.strm.Streak.Activate().Hit()
		activeGenerated.strm.CMD = exec.Command("tail", "-f", "/dev/null")
		go func() {
			activeGenerated.strm.CMD.Run()
		}()
		ctrls.streams = map[string]*streaming.Stream{
			generated.dirPath:       &generated.strm,
			activeGenerated.dirPath: &activeGenerated.strm,
		}
		<-time.After(time.Second * 2)
		ctrls.cleanUnused()
		wg.Wait()
		assert.False(t, generated.strm.CMD.ProcessState.Success())
		assert.Nil(t, activeGenerated.strm.CMD.ProcessState)
	})

	t.Run("Should be able to clean everything if it is required", func(t *testing.T) {
		ctrls := NewController(cfg, fileServer)
		wg := &sync.WaitGroup{}
		wg.Add(2)

		generated := generateStream(nil, "")
		generated.strm.Streak.Deactivate()
		generated.strm.CMD = exec.Command("tail", "-f", "/dev/null")
		go func() {
			generated.strm.CMD.Run()
			wg.Done()
		}()
		activeGenerated := generateStream(nil, "")
		activeGenerated.strm.Streak.Activate().Hit()
		activeGenerated.strm.CMD = exec.Command("tail", "-f", "/dev/null")
		go func() {
			activeGenerated.strm.CMD.Run()
			wg.Done()
		}()
		ctrls.streams = map[string]*streaming.Stream{
			generated.dirPath:       &generated.strm,
			activeGenerated.dirPath: &activeGenerated.strm,
		}
		<-time.After(time.Second * 2)
		ctrls.cleanUp()
		wg.Wait()
		assert.False(t, generated.strm.CMD.ProcessState.Success())
		assert.False(t, activeGenerated.strm.CMD.ProcessState.Success())
	})

	t.Run("Should be able to serve required files for known streams", func(t *testing.T) {
		storeDir := "./test"
		assert.Nil(t, os.MkdirAll(storeDir, os.ModePerm))
		localFileserver := http.FileServer(http.Dir(storeDir))
		ctrls := NewController(cfg, localFileserver)
		router := httprouter.New()
		router.POST("/start", ctrls.StartStreamHandler)
		router.GET("/stream/*filepath", ctrls.FileHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		generated := generateStream(nil, "")
		generated.strm.Streak.Activate().Hit()
		ctrls.streams = map[string]*streaming.Stream{
			generated.dirPath: &generated.strm,
		}

		assert.Nil(t, os.MkdirAll(fmt.Sprintf("%s/%s", storeDir, generated.dirPath), os.ModePerm))
		file, err := os.Create(fmt.Sprintf("%s/%s/index.m3u8", storeDir, generated.dirPath))
		assert.Nil(t, err)
		testString := gofakeit.BS()
		file.WriteString(testString)
		assert.Nil(t, file.Close())

		res, err := http.Get(fmt.Sprintf("%s/stream/%s/index.m3u8", server.URL, generated.dirPath))
		assert.Nil(t, err)
		b, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		assert.Equal(t, testString, string(b))
		assert.Nil(t, os.RemoveAll(storeDir))
	})

	t.Run("Should be able to restart stopped streams even with files", func(t *testing.T) {
		storeDir := "./test"
		assert.Nil(t, os.MkdirAll(storeDir, os.ModePerm))
		localFileserver := http.FileServer(http.Dir(storeDir))
		ctrls := NewController(cfg, localFileserver)
		ctrls.processor = mockProcessor{}
		router := httprouter.New()
		router.POST("/start", ctrls.StartStreamHandler)
		router.GET("/stream/*filepath", ctrls.FileHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		generated := generateStream(nil, "")
		generated.strm.Streak.Deactivate()
		ctrls.streams = map[string]*streaming.Stream{
			generated.dirPath: &generated.strm,
		}

		assert.Nil(t, os.MkdirAll(fmt.Sprintf("%s/%s", storeDir, generated.dirPath), os.ModePerm))
		file, err := os.Create(fmt.Sprintf("%s/%s/index.m3u8", storeDir, generated.dirPath))
		assert.Nil(t, err)
		testString := gofakeit.BS()
		file.WriteString(testString)
		assert.Nil(t, file.Close())

		res, err := http.Get(fmt.Sprintf("%s/stream/%s/index.m3u8", server.URL, generated.dirPath))
		assert.Nil(t, err)
		b, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		assert.Equal(t, testString, string(b))

		strm := ctrls.streams[generated.dirPath]
		assert.True(t, strm.Streak.IsActive())
		assert.Nil(t, os.RemoveAll(storeDir))
	})
}
