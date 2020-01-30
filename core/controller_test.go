package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIDByPath(t *testing.T) {
	var ctrl *Controller
	tt := []struct {
		Input  string
		Output string
	}{
		{
			Input:  "/73f23853-f041-4cce-84fa-c237b9b0be92/index.m3u8",
			Output: "73f23853-f041-4cce-84fa-c237b9b0be92",
		},
		{
			Input:  "/73f23853-f041-4cce-84fa-c237b9b0be92/1.ts",
			Output: "73f23853-f041-4cce-84fa-c237b9b0be92",
		},
	}

	for i, testCase := range tt {
		if !assert.Equal(t, testCase.Output, ctrl.getIDByPath(testCase.Input)) {
			t.Error(fmt.Errorf("%d testcase is failing", i))
		}
	}
}
