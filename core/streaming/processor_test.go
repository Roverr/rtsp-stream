package streaming

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetURIDirectory(t *testing.T) {
	tt := []struct {
		Input  string
		Output string
		Err    error
	}{
		{
			Input:  `<IMG """><SCRIPT>alert("XSS")</SCRIPT>">`,
			Output: "",
			Err:    ErrInvalidHost,
		},
		{
			Input:  "rtsp://test:test@192.168.0.1/Streaming/Channels/101",
			Output: "192-168-0-1-streaming-channels-101",
			Err:    nil,
		},
		{
			Input:  "rtsp://test:test@../../../Streaming/(SELECT * from users)",
			Output: "-streaming-select-from-users",
			Err:    nil,
		},
		{
			Input:  "rtsp://test:test@../../../etc/ssl",
			Output: "-etc-ssl",
			Err:    nil,
		},
		{
			Input:  "rtsp://test:test@127.0.0.1/I am a long url's_-?ASDF@£$%£%^é.html",
			Output: "127-0-0-1-i-am-a-long-urls-",
			Err:    nil,
		},
		{
			Input:  fmt.Sprintf("rtsp://test:test@127.0.0.1/%s", `<a href="/" alt="Fab.com | Aqua Paper Map 22"" title="Fab.com | Aqua Paper Map 22" - fab.com">test</a>`),
			Output: "127-0-0-1-a-href-alt-fab-com-aqua-paper-map-22-title-fab-com-aqua-paper-map-22-fab-comtest-a",
			Err:    nil,
		},
	}

	for i, testCase := range tt {
		URL, err := GetURIDirectory(testCase.Input)
		if !assert.Equal(t, testCase.Err, err) || !assert.Equal(t, testCase.Output, URL) {
			t.Error(fmt.Errorf("%d testcase is failing for TestGetURIDirectory", i))
		}
	}

}

func TestCreateDirectories(t *testing.T) {
	tt := []struct {
		Input   string
		DirPath string
		NewPath string
		Err     error
	}{
		{
			Input:   `<IMG """><SCRIPT>alert("XSS")</SCRIPT>">`,
			DirPath: "",
			NewPath: "",
			Err:     ErrInvalidHost,
		},
		{
			Input:   "rtsp://test:test@192.168.0.1/Streaming/Channels/101",
			DirPath: "192-168-0-1-streaming-channels-101",
			NewPath: "test/192-168-0-1-streaming-channels-101",
			Err:     nil,
		},
		{
			Input:   "rtsp://test:test@../../../etc/ssl",
			DirPath: "-etc-ssl",
			NewPath: "test/-etc-ssl",
			Err:     nil,
		},
		{
			Input:   "rtsp://test:test@127.0.0.1/I am a long url's_-?ASDF@£$%£%^é.html",
			DirPath: "127-0-0-1-i-am-a-long-urls-",
			NewPath: "test/127-0-0-1-i-am-a-long-urls-",
			Err:     nil,
		},
	}
	storeDir := "./test"
	for i, testCase := range tt {

		dirPath, newPath, err := createDirectoryForURI(testCase.Input, storeDir)
		if !assert.Equal(t, testCase.Err, err) || !assert.Equal(t, testCase.DirPath, dirPath) || !assert.Equal(t, testCase.NewPath, newPath) {
			t.Error(fmt.Errorf("%d testcase is failing for TestCreateDirectories", i))
		}
	}

	assert.Nil(t, os.RemoveAll(storeDir))
}
