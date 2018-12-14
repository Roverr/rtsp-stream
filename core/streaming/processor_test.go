package streaming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetURIDirectory(t *testing.T) {
	URL, err := GetURIDirectory("rtsp://test:test@192.168.0.1/Streaming/Channels/101")
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.1-streaming-channels-101", URL)
}
