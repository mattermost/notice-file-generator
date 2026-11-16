package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNpmLoadSuccess(t *testing.T) {
	dep := Dependency{Name: "mattermost-client"}
	err := dep.NpmLoad()

	assert.Nil(t, err)
	assert.Equal(t, "Andy Lo-A-Foe", dep.Author.Name)
	assert.Equal(t, "Mattermost client", dep.Description)
	assert.Equal(t, "https://github.com/loafoe/mattermost-client#readme", dep.HomePage)
	assert.Equal(t, "MIT", dep.License)
	assert.Equal(t, "mattermost-client", dep.Name)
	assert.Equal(t, "git", dep.Repository.Type)
	assert.Equal(t, "git+https://github.com/loafoe/mattermost-client.git", dep.Repository.URL)
}

func TestNpmLoadFailure(t *testing.T) {
	dep := Dependency{Name: "invalid-npm-package"}
	err := dep.NpmLoad()
	assert.NotNil(t, err)
}

func TestPopulateLicenseByHomePage(t *testing.T) {
	dep := Dependency{Name: "mattermost-client", HomePage: "https://github.com/loafoe/mattermost-client#readme"}

	license := dep.PopulateLicence()

	assert.NotEmpty(t, license)
}

func TestPopulateLicenseByUrl(t *testing.T) {
	dep := Dependency{Name: "mattermost-client", Repository: DependencyRepository{URL: "https://github.com/loafoe/mattermost-client"}}

	license := dep.PopulateLicence()

	assert.NotEmpty(t, license)
}

func TestNpmPackageName(t *testing.T) {
	tests := []struct {
		name, spec, want string
	}{
		{"@typescript/native", "npm:typescript@7.0.2", "typescript"},
		{"typescript", "npm:@typescript/typescript6@6.0.2", "@typescript/typescript6"},
		{"local", "npm:pkg", "pkg"},
		{"local", "npm:@scope/pkg", "@scope/pkg"},
		{"local", "npm:", "local"},
		{"react", "19.3.1", "react"},
		{"marked", "github:mattermost/marked#467bc9e", "marked"},
		{"@mattermost/rnutils", "file:./libraries/@mattermost/rnutils", "@mattermost/rnutils"},
	}
	for _, tt := range tests {
		got, err := npmPackageName(tt.name, tt.spec)
		assert.NoError(t, err, tt.spec)
		assert.Equal(t, tt.want, got, tt.spec)
	}
}

func TestNpmPackageNameRejectsInvalidAlias(t *testing.T) {
	for _, spec := range []string{
		"npm:../../evil@1.0.0",
		"npm:pkg?x=1@1.0.0",
		"npm:Upper@1.0.0",
		"npm:@scope@1.0.0",
		"npm:.hidden@1.0.0",
		"npm:" + strings.Repeat("a", 215),
	} {
		_, err := npmPackageName("local", spec)
		assert.Error(t, err, spec)
	}
}

func TestPopulateJSDependenciesResolvesAliases(t *testing.T) {
	config := &Config{IncludeDevDependencies: true}

	deps, err := config.PopulateJSDependencies("testdata/npm-alias/package.json")
	assert.NoError(t, err)

	var names []string
	for _, d := range deps {
		names = append(names, d.Name)
	}
	assert.ElementsMatch(t, []string{"typescript", "@typescript/typescript6", "react", "@mattermost/commonmark"}, names)
}

func TestIgnoreDependencies(t *testing.T) {
	os.Args = append(os.Args, "-p=/tmp/test", "-t=token", "-c=testdata/dependency_test.yaml")

	config := newConfig()
	allDeps, err := PopulateDependencies(config)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(allDeps))
	assert.Equal(t, "wix", allDeps[0].Name)

	os.Args = os.Args[:len(os.Args)-3]
}
