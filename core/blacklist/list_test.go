package blacklist

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	url := "rtsp://user:password@host.com/Channels/202"
	list := NewList(time.Hour*1, 2)
	list.AddOrIncrease(url)
	assert.Equal(t, false, list.IsBanned(url))
	list.AddOrIncrease(url)
	assert.Equal(t, false, list.IsBanned(url))
	list.AddOrIncrease(url)
	list.AddOrIncrease(url)
	assert.Equal(t, true, list.IsBanned(url))
	record, ok := list.list.Load(url)
	assert.Equal(t, true, ok)
	assert.Equal(t, true, record.(IRecord).GetBanTime().Before(time.Now().Add(time.Hour*1+1)))
}
