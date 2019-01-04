package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestController(t *testing.T) {
	cfg := config.InitConfig()
	fileServer := http.FileServer(http.Dir(cfg.StoreDir))
	ctrls := NewController(cfg, fileServer)
	router := httprouter.New()
	router.GET("/list", ctrls.ListStreamHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	res, err := http.Get(fmt.Sprintf("%s/list", server.URL))
	assert.Nil(t, err)
	b, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	var result []streamDto
	assert.Nil(t, json.Unmarshal(b, &result))
}
