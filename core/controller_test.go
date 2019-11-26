package core

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/Roverr/hotstreak"
	"github.com/google/uuid"

	"github.com/Roverr/rtsp-stream/core/streaming"
	"github.com/brianvoe/gofakeit"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())
	os.Setenv("AUTH_JWT_ENABLED", "false")
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
	streak := hs
	if hs == nil {
		streak = hotstreak.New(hotstreak.Config{
			Limit:      10,
			HotWait:    time.Minute * 3,
			ActiveWait: time.Minute * 4,
		})
	}
	id := uuid.New().String()
	return generatedStream{
		strm: streaming.Stream{
			Mux:         &sync.RWMutex{},
			Path:        fmt.Sprintf("/stream/%s/index.m3u8", id),
			OriginalURI: uri,
			Streak:      streak,
		},
		dirPath: id,
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

func (m mockManager) WaitForStream(path string) chan bool {
	if m.instead != nil {
		return (*m.instead)(path)
	}
	streamResolved := make(chan bool, 1)
	streamResolved <- m.resolve
	return streamResolved
}

type mockProcessor struct{}

var _ streaming.IProcessor = (*mockProcessor)(nil)

func (m mockProcessor) NewProcess(path, URI string) *exec.Cmd {
	return nil
}

func (m mockProcessor) NewStream(URI string) (*streaming.Stream, string, string) {
	generated := generateStream(nil, URI)
	return &generated.strm, generated.strm.Path, generated.dirPath
}

func (m mockProcessor) Restart(stream *streaming.Stream, path string) error {
	stream.Streak.Activate().Hit()
	return nil
}

/*
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

	t.Run("Should be blocked if auth is on and no token available", func(t *testing.T) {
		conf := config.InitConfig()
		conf.JWTEnabled = true
		ctrls := NewController(conf, fileServer)
		router := httprouter.New()
		router.GET("/list", ctrls.ListStreamHandler)
		router.POST("/start", ctrls.StartStreamHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		res, err := http.Get(fmt.Sprintf("%s/list", server.URL))
		assert.Nil(t, err)
		assert.Equal(t, http.StatusForbidden, res.StatusCode)
	})

	t.Run("Should get list back if authenticated", func(t *testing.T) {
		conf := config.InitConfig()
		conf.JWTEnabled = true
		ctrls := NewController(conf, fileServer)
		router := httprouter.New()
		router.GET("/list", ctrls.ListStreamHandler)
		router.POST("/start", ctrls.StartStreamHandler)
		server := httptest.NewServer(router)
		defer server.Close()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{})
		tokenString, err := token.SignedString([]byte("macilaci"))
		assert.Nil(t, err)
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/list", server.URL), nil)
		assert.Nil(t, err)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))

		res, err := (&http.Client{}).Do(req)
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
		ctrls.index[generated.strm.OriginalURI] = generated.dirPath
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
		b, err := json.Marshal(dto)
		assert.Nil(t, err)
		res, err := http.Post(fmt.Sprintf("%s/start", server.URL), "application/json", bytes.NewBuffer(b))
		assert.Nil(t, err)
		b, err = ioutil.ReadAll(res.Body)
		assert.Nil(t, err)
		var result StreamDto
		assert.Nil(t, json.Unmarshal(b, &result))
		index, ok := ctrls.index[dto.URI]
		assert.True(t, ok)
		strm, ok := ctrls.streams[index]
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
		ctrls.index[generated.strm.OriginalURI] = generated.dirPath
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
		index, ok := ctrls.index[dto.URI]
		assert.True(t, ok)
		strm, ok := ctrls.streams[index]
		assert.True(t, ok)
		assert.Equal(t, result.URI, strm.Path)
		assert.True(t, ctrls.streams[generated.dirPath].Streak.IsActive())
	})
	t.Run("Should be able to receive unexpected error if something happens", func(t *testing.T) {
		ctrls := NewController(cfg, fileServer)
		ctrls.manager = mockManager{resolve: false}
		ctrls.processor = mockProcessor{}
		ctrls.streams = map[string]*streaming.Stream{}
		ctrls.index = map[string]string{}
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
		ctrls.index = map[string]string{}
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
		ctrls.index = map[string]string{
			generated.strm.OriginalURI:       generated.dirPath,
			activeGenerated.strm.OriginalURI: activeGenerated.dirPath,
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
		ctrls.index = map[string]string{
			generated.strm.OriginalURI:       generated.dirPath,
			activeGenerated.strm.OriginalURI: activeGenerated.dirPath,
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
		ctrls.index = map[string]string{
			generated.strm.OriginalURI: generated.dirPath,
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
		ctrls.index = map[string]string{
			generated.strm.OriginalURI: generated.dirPath,
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
*/
