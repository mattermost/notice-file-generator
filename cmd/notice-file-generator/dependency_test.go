package main

import (
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
