package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Title                  string   `yaml:"title"`
	Copyright              string   `yaml:"copyright"`
	Description            string   `yaml:"description"`
	Reviewers              []string `yaml:"reviewers"`
	Search                 []string `yaml:"search"`
	IncludeDevDependencies bool     `yaml:"includeDevDependencies"`
	AdditionalDependencies []string `yaml:"additionalDependencies"`
	IgnoreDependencies     []string `yaml:"ignoreDependencies"`
	Name                   string   `yaml:"-"`
	Path                   string   `yaml:"-"`
	GHToken                string   `yaml:"-"`
	GoFiles                []string `yaml:"-"`
	JSFIles                []string `yaml:"-"`
}

type Argument struct {
	Name         string
	DefaultValue string
	Description  string
}

func (c *Config) NoticeDirPath() string {
	return fmt.Sprintf("%s/.notice", c.Path)
}

func (c *Config) NoticeWorkPath() string {
	return fmt.Sprintf("%s/.notice-work", c.Path)
}

func (c *Config) NoticeFilePath() string {
	return fmt.Sprintf("%s/NOTICE.txt", c.Path)
}

func (c *Config) determineRepoFiles() {
	for _, search := range c.Search {

		if strings.Contains(search, "package.json") {
			JSFile, _ := filepath.Abs(filepath.Join(c.Path, search))
			c.JSFIles = append(c.JSFIles, JSFile)
		}

		if strings.Contains(search, "go.mod") {
			goFile, _ := filepath.Abs(filepath.Join(c.Path, search))
			c.GoFiles = append(c.GoFiles, goFile)
		}

	}
}

func getArgs() map[string]string {
	m := make(map[string]string)
	supportedArguments := []Argument{
		{"p", ".", "Repository Path"},
		{"t", "", "Github Authentication Token"},
		{"c", "", "Configuration File Path"},
	}
	flagsDefined := (flag.Lookup(supportedArguments[0].Name) != nil)

	if !flagsDefined {
		// Define the flags before parsing the CLI args
		for _, arg := range supportedArguments {
			_ = flag.String(arg.Name, arg.DefaultValue, arg.Description)
		}
	}
	// The flags are now defined: parse their value, then retrieve it
	flag.Parse()
	for _, arg := range supportedArguments {
		m[arg.Name] = (flag.Lookup(arg.Name).Value.String())
	}
	return m
}

func newConfig() *Config {
	args := getArgs()
	repositoryPath := args["p"]
	githubToken := args["t"]
	configFilePath := args["c"]

	if len(configFilePath) == 0 || len(repositoryPath) == 0 {
		fmt.Println("Usage: main.go -p path -t token -c configFile")
		flag.PrintDefaults()
		os.Exit(1)
	}

	content, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("%s - Configuration file error! %v", repositoryPath, err)
	}
	// Path always exist, no need to check error
	repoFullPath, _ := filepath.Abs(repositoryPath)
	config := &Config{
		Path:    repoFullPath,
		GHToken: githubToken,
	}

	if err = yaml.Unmarshal(content, config); err != nil {
		log.Fatalf("%s - Configuration file error! %v", repositoryPath, err)
	}

	config.determineRepoFiles()
	return config

}
