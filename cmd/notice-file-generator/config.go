package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
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

func (c *Config) IsJsRepo() bool {
	for _, search := range c.Search {
		if strings.Contains(search, "package.json") {
			return true
		}
	}

	return false
}

func (c *Config) IsGoRepo() bool {
	for _, search := range c.Search {
		if strings.Contains(search, "go.mod") {
			return true
		}
	}

	return false
}

func (c *Config) IsPythonRepo() bool {
	for _, search := range c.Search {
		if strings.Contains(search, "Pipfile") {
			return true
		}
	}

	return false
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

	return config

}
