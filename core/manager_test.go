package core

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestManager(t *testing.T) {
	storeDir := "./test"
	assert.Nil(t, os.MkdirAll(storeDir, os.ModePerm))

	t.Run("Should return with true, because the file is created", func(t *testing.T) {
		mngr := NewManager(time.Second * 15)
		id := uuid.New().String()
		dirPath := fmt.Sprintf("%s/%s", storeDir, id)
		physicalPath := fmt.Sprintf("%s/index.m3u8", dirPath)
		assert.Nil(t, os.MkdirAll(dirPath, os.ModePerm))
		cmd := exec.Command("touch", physicalPath)
		streamResolved := mngr.Start(cmd, physicalPath)
		success := <-streamResolved
		assert.True(t, success)
	})

	t.Run("Should return with false, the process errors out before the file is created", func(t *testing.T) {
		mngr := NewManager(time.Second * 15)
		id := uuid.New().String()
		dirPath := fmt.Sprintf("%s/%s", storeDir, id)
		physicalPath := fmt.Sprintf("%s/index.m3u8", dirPath)
		assert.Nil(t, os.MkdirAll(dirPath, os.ModePerm))
		cmd := exec.Command("exit", "1")
		streamResolved := mngr.Start(cmd, physicalPath)
		success := <-streamResolved
		assert.False(t, success)
	})

	t.Run("Should return with false, if the process just times out", func(t *testing.T) {
		mngr := NewManager(time.Second * 2)
		id := uuid.New().String()
		dirPath := fmt.Sprintf("%s/%s", storeDir, id)
		physicalPath := fmt.Sprintf("%s/index.m3u8", dirPath)
		assert.Nil(t, os.MkdirAll(dirPath, os.ModePerm))
		cmd := exec.Command("sleep", "20")
		streamResolved := mngr.Start(cmd, physicalPath)
		success := <-streamResolved
		assert.False(t, success)
	})

	assert.Nil(t, os.RemoveAll(storeDir))
}
