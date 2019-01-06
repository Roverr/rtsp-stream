package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineHost(t *testing.T) {
	tt := []struct {
		Input  string
		Output string
	}{
		{
			Input:  "/192-168-0-1-channels-robinsion-001/index.m3u8",
			Output: "192-168-0-1-channels-robinsion-001",
		},
		{
			Input:  "/192-168-0-1-channels-wwww-001/asd.jpeg",
			Output: "192-168-0-1-channels-wwww-001",
		},
	}

	for i, testCase := range tt {
		if !assert.Equal(t, testCase.Output, determineHost(testCase.Input)) {
			t.Error(fmt.Errorf("%d testcase is failing", i))
		}
	}
}
