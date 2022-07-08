package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type RepoType int

const (
	JsRepo RepoType = iota
	GoRepo
	PythonRepo
)

type Config struct {
	Title           string   `yaml:"title"`
	Copyright       string   `yaml:"copyright"`
	Description     string   `yaml:"description"`
	Reviewers       []string `yaml:"reviewers"`
	Search          []string `yaml:"search"`
	Dependencies    []string `yaml:"dependencies"`
	DevDependencies []string `yaml:"devDependencies"`
	Name            string   `yaml:"-"`
	Path            string   `yaml:"-"`
	GHToken         string   `yaml:"-"`
	RepoType        RepoType `yaml:"-"`
}

func (c *Config) NoticeDirPath() string {
	return fmt.Sprintf("%s-notice", c.Path)
}

func (c *Config) NoticeWorkPath() string {
	return fmt.Sprintf("%s-notice-work", c.Path)
}

func (c *Config) NoticeFilePath() string {
	return fmt.Sprintf("%s/NOTICE.txt", c.Path)
}

func newConfig() *Config {
	repositoryPath := flag.String("p", "", "Repository Path")
	repositoryName := flag.String("n", "", "Name of the Repo")
	githubToken := flag.String("t", "", "Github Authentication Token")
	configFilePath := flag.String("c", "", "Configuration File Path")

	flag.Parse()

	if len(*repositoryPath) == 0 || len(*repositoryName) == 0 || len(*githubToken) == 0 || len(*configFilePath) == 0 {
		fmt.Println("Usage: main.go -n name -p path -t token -c configFile")
		flag.PrintDefaults()
		os.Exit(1)
	}

	content, err := os.ReadFile(*configFilePath)
	if err != nil {
		log.Fatalf("%s - Configuration file error! %v", *repositoryPath, err)
	}
	config := &Config{
		Name:    *repositoryName,
		Path:    *repositoryPath,
		GHToken: *githubToken,
	}
	if err = yaml.Unmarshal(content, config); err != nil {
		log.Fatalf("%s - Configuration file error! %v", *repositoryPath, err)
	}

	for _, search := range config.Search {
		if strings.Contains(search, "package.json") {
			config.RepoType = JsRepo
			break
		}
		if strings.Contains(search, "go.mod") {
			config.RepoType = GoRepo
			break
		}
		if strings.Contains(search, "Pipfile") {
			config.RepoType = PythonRepo
			break
		}
	}

	return config

}
