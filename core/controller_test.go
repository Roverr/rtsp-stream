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

func TestRedirectAlias(t *testing.T) {
	c := &Controller{
		alias: map[string]string{
			"alias": "id",
		},
	}

	tt := []struct {
		FilePath string
		Expected string
	}{
		{
			FilePath: "stream/12345/index.m3u8",
			Expected: "",
		},
		{
			FilePath: "stream/12345/22.ts",
			Expected: "",
		},
		{
			FilePath: "stream/alias/22.ts",
			Expected: "/stream/id/22.ts",
		},
		{
			FilePath: "stream/alias/index.m3u8",
			Expected: "/stream/id/index.m3u8",
		},
	}

	for i, testCase := range tt {
		id := c.getIDByPath(testCase.FilePath)

		url, redirect := c.shouldRedirectAlias(id, testCase.FilePath)

		if !assert.Equal(t, testCase.Expected, url, testCase.FilePath) || !assert.Equal(t, testCase.Expected != "", redirect) {
			t.Error(fmt.Errorf("%d testcase is failing", i))
		}
	}
}
