package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoticeDirPath(t *testing.T) {
	config := Config{Path: "/tmp/work"}
	assert.Equal(t, "/tmp/work-notice", config.NoticeDirPath())
}

func TestNoticeWorkPath(t *testing.T) {
	config := Config{Path: "/tmp/work"}
	assert.Equal(t, "/tmp/work-notice-work", config.NoticeWorkPath())
}

func TestNoticeFilePath(t *testing.T) {
	config := Config{Path: "/tmp/work"}
	assert.Equal(t, "/tmp/work/NOTICE.txt", config.NoticeFilePath())
}
func TestIsJsRepo(t *testing.T) {
	jsconfig := Config{Search: []string{"", "package.json"}}
	goconfig := Config{Search: []string{"", "go.mod"}}
	assert.True(t, jsconfig.IsJsRepo())
	assert.False(t, goconfig.IsJsRepo())
}

func TestNewConfig(t *testing.T) {
	os.Args = append(os.Args, "-n=test", "-p=/tmp/test", "-t=token", "-c=testdata/test.yaml")

	config := newConfig()
	assert.Equal(t, "test", config.Name)
	assert.Equal(t, "/tmp/test", config.Path)
	assert.Equal(t, "token", config.GHToken)
	assert.Equal(t, "Notice Title", config.Title)
	assert.Equal(t, "Notice Copyright", config.Copyright)
	assert.Equal(t, "Notice Description", config.Description)
	assert.Equal(t, 1, len(config.Reviewers))
	assert.Equal(t, "reviewerOne", config.Reviewers[0])
	assert.Equal(t, 1, len(config.Search))
	assert.Equal(t, "package.json", config.Search[0])
	assert.Equal(t, 1, len(config.Dependencies))
	assert.Equal(t, "wix", config.Dependencies[0])
	assert.Equal(t, 1, len(config.DevDependencies))
	assert.Equal(t, "webpack", config.DevDependencies[0])
}
