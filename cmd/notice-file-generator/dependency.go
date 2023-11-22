package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/google/go-github/github"
	"golang.org/x/mod/modfile"
	"golang.org/x/oauth2"
)

type DependencyType int

const (
	JsDep DependencyType = iota
	GoDep
	PythonDep
)

type NpmPackage struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type Dependencies struct {
	sync.Mutex
	value []Dependency
}

func (deps *Dependencies) append(x Dependency) {
	defer deps.Unlock()
	deps.Lock()
	deps.value = append(deps.value, x)
}

type Dependency struct {
	Name           string               `json:"name"`
	FullName       string               `json:"-"`
	Description    string               `json:"description"`
	Author         DependencyAuthor     `json:"author"`
	License        string               `json:"license"`
	Repository     DependencyRepository `json:"repository"`
	HomePage       string               `json:"homepage"`
	DependencyType DependencyType       `json:"-"`
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

var regexPythonDep = regexp.MustCompile(`^#\s*[R|r]epo(sitory)*:\s*(?P<url>https:\/\/github.com\/(?P<full_name>.*\/(?P<name>.*)))`)

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
	data, err := HTTPGet("https://registry.npmjs.org/" + d.Name)

	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), &d); err != nil {
		if ue, ok := err.(*json.UnmarshalTypeError); ok {
			switch ue.Field {
			case "Author":
				d.Author = DependencyAuthor{
					Name:  ue.Value,
					Email: "",
				}
			case "Repository":
				d.Repository = DependencyRepository{
					Type: "",
					URL:  ue.Value,
				}
			}
			return nil
		} else {
			return err
		}

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

		if len(repo.GetHomepage()) == 0 {
			d.HomePage = repo.GetHTMLURL()
		} else {
			d.HomePage = repo.GetHomepage()
		}

		if repo.GetLicense() != nil {
			d.License = repo.License.GetName()
		} else {
			log.Printf("There is no licence available for %s", d.Name)
		}
	}
	return nil
}

func (d *Dependency) Generate(config *Config) error {
	filename := GenerateFileName(d.Name)

	if err := MoveExistingNotice(config, filename); err != nil {
		switch d.DependencyType {
		case JsDep:
			log.Printf("Generating notice for %s npm dependency from NPM registry", d.Name)
			if err = d.NpmLoad(); err != nil {
				log.Printf("NPM load failed  %s", d.Name)
				return err
			}
		case GoDep:
			log.Printf("Generating notice for %s go.mod dependency from Github", d.Name)
			if err = d.LoadFromGithub(config); err != nil {
				log.Printf("GitHub load failed  %s", d.Name)
				return err
			}
			// case PythonDep:
			// 	if err = d.LoadFromGithub(config); err != nil {
			// 		log.Printf("GitHub load failed  %s", d.Name)
			// 		return err
			// 	}
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
		log.Printf("Using existing notice for %s dependency", d.Name)
	}
	return nil
}

func (d *Dependency) Load(config *Config) string {
	filename := GenerateFileName(d.Name)
	c, _ := os.ReadFile(config.NoticeWorkPath() + "/" + filename)
	return string(c)
}

func PopulateJSDependencies(packageJSON string) ([]Dependency, error) {
	var dependencies []string

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
	// for _, dependency := range config.Dependencies {
	// 	if IndexOf(dependencies, dependency) == -1 {
	// 		dependencies = append(dependencies, dependency)
	// 	}
	// }
	deps := []Dependency{}

	for _, dependency := range dependencies {
		deps = append(deps, Dependency{Name: dependency, DependencyType: JsDep})
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

func PopulateGoDependencies(goModFile string) ([]Dependency, error) {
	dependencies := Dependencies{}
	dependencies.append(Dependency{
		Name:           "Go",
		FullName:       "github.com/golang/go",
		HomePage:       "https://go.dev/",
		Description:    "The Go programming language",
		Author:         DependencyAuthor{Name: "The Go authors"},
		License:        "BSD-style",
		DependencyType: GoDep,
		Repository: DependencyRepository{
			Type: "git",
			URL:  "github.com/golang/go",
		},
	})
	o, e := os.ReadFile(goModFile)
	if e != nil {
		log.Fatalf("Could not read go.mod file at location %s . Error:  %v", goModFile, e)
	}

	f, err := modfile.Parse("go.mod", o, nil)
	if err != nil {
		log.Fatalf("Invalid go.mod file. %v", err)
	}

	var wg sync.WaitGroup

	for _, r := range f.Require {
		wg.Add(1)
		r := r

		go func() {
			defer wg.Done()

			log.Printf("Populating %s go.mod dependency", r.Mod.String())
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
					dependencies.append(Dependency{
						Name:           name,
						FullName:       gi.ImportPrefix,
						DependencyType: GoDep,
						Repository: DependencyRepository{
							Type: gi.Vcs,
							URL:  gi.RepoRoot,
						},
					})
				}
			}
		}()
	}

	wg.Wait()

	return dependencies.value, nil
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
		if regexPythonDep.MatchString(line) {
			matches := regexPythonDep.FindStringSubmatch(line)
			dependencies = append(dependencies, Dependency{
				Name:     matches[regexPythonDep.SubexpIndex("name")],
				FullName: matches[regexPythonDep.SubexpIndex("full_name")],
				Repository: DependencyRepository{
					Type: "https",
					URL:  matches[regexPythonDep.SubexpIndex("url")],
				},
			})
		}
	}
	return dependencies, nil
}

func PopulateDependencies(config *Config) ([]Dependency, error) {
	var deps []Dependency

	for _, modFile := range config.GoFiles {
		d, err := PopulateGoDependencies(modFile)
		if err != nil {
			return deps, err
		}
		deps = append(deps, d...)
	}

	for _, jsFile := range config.JSFIles {
		d, err := PopulateJSDependencies(jsFile)
		if err != nil {
			return deps, err
		}
		deps = append(deps, d...)
	}

	// if config.DependencyType == PythonDep {
	// 	return PopulatePythonDependencies(config)
	// }

	return deps, nil
}
