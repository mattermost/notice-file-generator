package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoticeDirPath(t *testing.T) {
	config := Config{Path: "/tmp/work"}
	assert.Equal(t, "/tmp/work/.notice", config.NoticeDirPath())
}

func TestNoticeWorkPath(t *testing.T) {
	config := Config{Path: "/tmp/work"}
	assert.Equal(t, "/tmp/work/.notice-work", config.NoticeWorkPath())
}

func TestNoticeFilePath(t *testing.T) {
	config := Config{Path: "/tmp/work"}
	assert.Equal(t, "/tmp/work/NOTICE.txt", config.NoticeFilePath())
}
func TestRepoFiles(t *testing.T) {
	path, _ := filepath.Abs(".")
	jsconfig := Config{Search: []string{"", "package.json"}, Path: path}
	goconfig := Config{Search: []string{"", "go.mod"}, Path: path}
	pythonconfig := Config{Search: []string{"", "Pipfile"}}
	jsconfig.determineRepoFiles()
	goconfig.determineRepoFiles()
	pythonconfig.determineRepoFiles()
	assert.Equal(t, []string{filepath.Join(jsconfig.Path, "package.json")}, jsconfig.JSFIles)
	assert.Equal(t, []string{filepath.Join(goconfig.Path, "go.mod")}, goconfig.GoFiles)
}

func TestNewConfig(t *testing.T) {
	os.Args = append(os.Args, "-p=/tmp/test", "-t=token", "-c=testdata/test.yaml")

	config := newConfig()
	assert.Equal(t, "/tmp/test", config.Path)
	assert.Equal(t, "token", config.GHToken)
	assert.Equal(t, "Notice Title", config.Title)
	assert.Equal(t, "Notice Copyright", config.Copyright)
	assert.Equal(t, "Notice Description", config.Description)
	assert.Equal(t, 1, len(config.Reviewers))
	assert.Equal(t, "reviewerOne", config.Reviewers[0])
	assert.Equal(t, 1, len(config.Search))
	assert.Equal(t, "package.json", config.Search[0])
	assert.Equal(t, 2, len(config.AdditionalDependencies))
	assert.Equal(t, "wix", config.AdditionalDependencies[0])
	assert.Equal(t, "ignored", config.AdditionalDependencies[1])
	assert.Equal(t, 1, len(config.IgnoreDependencies))
	assert.Equal(t, "ignored", config.IgnoreDependencies[0])

	os.Args = os.Args[:len(os.Args)-3]
}
