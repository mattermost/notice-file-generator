package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/google/go-github/github"
	"golang.org/x/mod/modfile"
	"golang.org/x/oauth2"
)

type NpmPackage struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type Dependency struct {
	Name        string               `json:"name"`
	FullName    string               `json:"-"`
	Description string               `json:"description"`
	Author      DependencyAuthor     `json:"author"`
	License     string               `json:"license"`
	Repository  DependencyRepository `json:"repository"`
	HomePage    string               `json:"homepage"`
}

type DependencyRepository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type DependencyAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

var regexpGoImport = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<\s*meta\s*name\s*=\s*"go-import"\s*content\s*=\s*"(?P<import_prefix>\S+)\s+(?P<vcs>\S+)\s+(?P<repo_root>\S+)"\s*/?>`),
	// source hut has the arguments the other way round
	regexp.MustCompile(`(?i)<\s*meta\s*content\s*=\s*"(?P<import_prefix>\S+)\s+(?P<vcs>\S+)\s+(?P<repo_root>\S+)"\s*name\s*=\s*"go-import"\s*/?>`),
}

var regexPythonRepo = regexp.MustCompile(`^#\s*[R|r]epo(sitory)*:\s*(?P<url>https:\/\/github.com\/(?P<full_name>.*\/(?P<name>.*)))`)

type GoImport struct {
	ImportPrefix string
	Vcs          string
	RepoRoot     string
}

func (d *Dependency) PopulateLicence() string {
	url := ""
	content := ""
	if d.HomePage != "" && strings.Contains(d.HomePage, "github.com") {
		url = d.HomePage
	} else if d.Repository.URL != "" {
		url = d.Repository.URL
	}
	if url != "" {
		prefix := ""
		parts := strings.FieldsFunc(url, func(r rune) bool {
			return r == '/' || r == '#' || r == '.'
		})

		for i := 0; i < len(parts); i++ {
			if parts[i] == "com" && parts[i-1] == "github" {
				prefix = fmt.Sprintf("%s/%s", parts[i+1], parts[i+2])
				break
			}
		}
		if prefix != "" {
			data, err := HTTPGet("https://raw.githubusercontent.com/" + prefix + "/HEAD/LICENSE.txt")
			if err != nil {
				data, err = HTTPGet("https://raw.githubusercontent.com/" + prefix + "/HEAD/LICENSE.md")
				if err != nil {
					data, err = HTTPGet("https://raw.githubusercontent.com/" + prefix + "/HEAD/LICENSE")
				}
			}
			if err == nil {
				content = data
			}
		}
	}
	return fmt.Sprintf("%s\n\n", content)
}

func (d *Dependency) NpmLoad() error {
	log.Printf("Load %s information by npm", d.Name)
	data, err := HTTPGet("https://registry.npmjs.org/" + d.Name)

	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), &d)
}

func (d *Dependency) LoadFromGithub(config *Config) error {
	if strings.Contains(d.Repository.URL, "github.com") {
		var gh = github.NewClient(nil)
		repoDef := strings.Split(d.Repository.URL, "/")
		scope := repoDef[len(repoDef)-2]
		repoName := strings.ReplaceAll(repoDef[len(repoDef)-1], ".git", "")

		if len(config.GHToken) != 0 {
			ctx := context.Background()
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: config.GHToken},
			)
			tc := oauth2.NewClient(ctx, ts)
			gh = github.NewClient(tc)
		}

		repo, _, err := gh.Repositories.Get(context.Background(), scope, repoName)
		if err != nil {
			log.Printf("Github Load Failed  %v", err)
			return err
		}
		if repo.Description != nil {
			d.Description = *repo.Description
		}
		authorName := *repo.Owner.Login
		author, _, err := gh.Users.Get(context.Background(), authorName)
		if err == nil {
			if author.Name != nil {
				authorName = *author.Name
			}
		}
		d.Author = DependencyAuthor{Name: authorName}

		if repo.Homepage != nil {
			d.HomePage = *repo.Homepage
		} else if repo.HTMLURL != nil {
			d.HomePage = *repo.HTMLURL
		}
		d.License = *repo.License.Name
	}
	return nil
}

func (d *Dependency) Generate(config *Config) error {
	filename := GenerateFileName(d.Name)

	if err := MoveExistingNotice(config, filename); err != nil {
		log.Printf("Populate notice of %s", d.Name)
		switch config.RepoType {
		case JsRepo:
			if err = d.NpmLoad(); err != nil {
				log.Printf("Npm load failed  %s", d.Name)
				return err
			}
		case GoRepo:
			if err = d.LoadFromGithub(config); err != nil {
				log.Printf("GitHub load failed  %s", d.Name)
				return err
			}
		case PythonRepo:
			if err = d.LoadFromGithub(config); err != nil {
				log.Printf("GitHub load failed  %s", d.Name)
				return err
			}
		}

		var out *os.File
		var writer *bufio.Writer

		out, err = os.OpenFile(config.NoticeWorkPath()+"/"+filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writer = bufio.NewWriter(out)

		if _, err = writer.WriteString(fmt.Sprintf("## %s\n\n", d.Name)); err != nil {
			log.Printf("Error while writing string %v", err)
		}
		if d.Author.Name != "" {
			if _, err = writer.WriteString(fmt.Sprintf("This product contains '%s' by %s.\n\n", d.Name, d.Author.Name)); err != nil {
				log.Printf("Error while writing string %v", err)
			}
		} else {
			if _, err = writer.WriteString(fmt.Sprintf("This product contains '%s'.\n\n", d.Name)); err != nil {
				log.Printf("Error while writing string %v", err)
			}
		}
		if d.Description != "" {
			if _, err = writer.WriteString(fmt.Sprintf("%s\n\n", d.Description)); err != nil {
				log.Printf("Error while writing string %v", err)
			}
		}
		if d.HomePage != "" {
			if _, err = writer.WriteString(fmt.Sprintf("* HOMEPAGE:\n  * %s\n\n", d.HomePage)); err != nil {
				log.Printf("Error while writing string %v", err)
			}
		}
		if d.License != "" {
			if _, err = writer.WriteString(fmt.Sprintf("* LICENSE: %s\n\n", d.License)); err != nil {
				log.Printf("Error while writing string %v", err)
			}
		}
		if _, err = writer.WriteString(d.PopulateLicence()); err != nil {
			log.Printf("Error while writing string %v", err)
		}
		writer.Flush()
		out.Close()

	} else {
		log.Printf("%s: Using existing notice txt.", d.Name)
	}
	return nil
}

func (d *Dependency) Load(config *Config) string {
	filename := GenerateFileName(d.Name)
	c, _ := os.ReadFile(config.NoticeWorkPath() + "/" + filename)
	return string(c)
}

func PopulateJSDependencies(config *Config) ([]Dependency, error) {
	var dependencies []string
	var packageJsons []string

	for _, searchPath := range config.Search {
		if strings.Contains(searchPath, "*") {
			matches, err := doublestar.Glob(filepath.Join(config.Path, searchPath))
			if err != nil {
				log.Printf("Failed to find any package.json for %s", searchPath)
			} else {
				packageJsons = append(packageJsons, matches...)
			}
		} else {
			packageJsons = append(packageJsons, fmt.Sprintf("%s/%s", config.Path, searchPath))
		}
	}
	for _, packageJSON := range packageJsons {
		o, err := os.ReadFile(packageJSON)
		if err != nil {
			log.Fatalf("%s-Invalid package json %v", packageJSON, err)
		}
		var npmPack NpmPackage

		if err := json.Unmarshal(o, &npmPack); err != nil {
			log.Fatalf("%s-Invalid package json %v", packageJSON, err)
		}

		for dependency := range npmPack.Dependencies {
			dependencies = append(dependencies, dependency)
		}
		for dependency := range npmPack.DevDependencies {
			dependencies = append(dependencies, dependency)
		}
	}
	for _, dependency := range config.Dependencies {
		if IndexOf(dependencies, dependency) == -1 {
			dependencies = append(dependencies, dependency)
		}
	}
	deps := []Dependency{}

	for _, dependency := range dependencies {
		deps = append(deps, Dependency{Name: dependency})
	}
	return deps, nil
}

func parseGoImport(data string) (GoImport, bool) {
	for _, r := range regexpGoImport {

		if !r.MatchString(data) {
			continue
		}

		matches := r.FindStringSubmatch(data)
		return GoImport{
			ImportPrefix: matches[r.SubexpIndex("import_prefix")],
			Vcs:          matches[r.SubexpIndex("vcs")],
			RepoRoot:     matches[r.SubexpIndex("repo_root")],
		}, true
	}

	return GoImport{}, false
}

func PopulateGoDependencies(config *Config) ([]Dependency, error) {
	dependencies := []Dependency{}
	dependencies = append(dependencies, Dependency{
		Name:        "Go",
		FullName:    "github.com/golang/go",
		HomePage:    "https://go.dev/",
		Description: "The Go programming language",
		Author:      DependencyAuthor{Name: "The Go authors"},
		License:     "BSD-style",
		Repository: DependencyRepository{
			Type: "git",
			URL:  "github.com/golang/go",
		},
	})
	o, _ := os.ReadFile(config.Path + "/go.mod")

	f, err := modfile.Parse("go.mod", o, nil)
	if err != nil {
		log.Fatalf("Invalid go.mod file. %v", err)
	}
	for _, r := range f.Require {
		if !r.Indirect {
			var data string
			var ok bool

			data, err = HTTPGet(fmt.Sprintf("https://%s?go-get=1", r.Mod.Path))
			if err != nil {
				parts := strings.Split(r.Mod.Path, "/")
				if len(parts) > 3 {
					moduleroot := strings.Join(parts[:3], "/")
					data, _ = HTTPGet(fmt.Sprintf("https://%s?go-get=1", moduleroot))
				}
			}

			gi, ok := parseGoImport(data)
			if !ok {
				log.Printf("unrecognised import %q (no go-import meta tags)", r.Mod.Path)
			} else {
				p := strings.Split(gi.ImportPrefix, "/")
				name := gi.ImportPrefix
				l := len(p)
				if l >= 2 {
					name = p[l-2] + "/" + p[l-1]
				}
				if strings.HasPrefix(gi.RepoRoot, "https://go.googlesource.com/") {
					parts := strings.Split(gi.RepoRoot, "/")
					gi.RepoRoot = fmt.Sprintf("https://github.com/golang/%s", parts[len(parts)-1])
				} else if strings.HasPrefix(name, "gopkg.in") {
					p := strings.Split(name, "/")
					p = strings.Split(p[1], ".")
					name = fmt.Sprintf("go-%s/%s", p[0], p[0])
					gi.RepoRoot = "https://github.com/" + name
				}
				dependencies = append(dependencies, Dependency{
					Name:     name,
					FullName: gi.ImportPrefix,
					Repository: DependencyRepository{
						Type: gi.Vcs,
						URL:  gi.RepoRoot,
					},
				})
			}
		}
	}
	return dependencies, nil
}

func PopulatePythonDependencies(config *Config) ([]Dependency, error) {
	dependencies := []Dependency{}
	inFile, err := os.Open(config.Path + "/Pipfile")
	if err != nil {
		log.Fatalf("Invalid pipfile. %v", err)
	}
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		if regexPythonRepo.MatchString(line) {
			matches := regexPythonRepo.FindStringSubmatch(line)
			dependencies = append(dependencies, Dependency{
				Name:     matches[regexPythonRepo.SubexpIndex("name")],
				FullName: matches[regexPythonRepo.SubexpIndex("full_name")],
				Repository: DependencyRepository{
					Type: "https",
					URL:  matches[regexPythonRepo.SubexpIndex("url")],
				},
			})
		}
	}
	return dependencies, nil
}

func PopulateDependencies(config *Config) ([]Dependency, error) {

	if config.RepoType == JsRepo {
		return PopulateJSDependencies(config)
	}

	if config.RepoType == GoRepo {
		return PopulateGoDependencies(config)
	}

	if config.RepoType == PythonRepo {
		return PopulatePythonDependencies(config)
	}

	return []Dependency{}, nil
}
