package updater

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type AppConfig struct {
	Docker *dockerConfig  `json:"docker,omitempty"`
	Build  []*buildConfig `json:"build,omitempty"`
}

type dockerConfig struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type buildConfig struct {
	Branch     string `json:"branch"`
	Arch       string `json:"arch"`
	Os         string `json:"os"`
	Dockerfile string `json:"dockerfile"`
}

func (config AppConfig) check() error {
	switch {
	case config.Docker == nil:
		return errors.New("config: 'docker' is missing")
	case config.Docker.User == "":
		return errors.New("config: 'docker > user' is empty")
	case config.Docker.Password == "":
		return errors.New("config: 'docker > password' is empty")
	case config.Build == nil:
		return errors.New("config: 'build' is missing")
	case len(config.Build) == 0:
		return errors.New("config: 'build' is empty")
	}
	for idx, build := range config.Build {
		switch {
		case build.Branch == "":
			return errors.New("config: 'build[" + string(idx) + "] > branch' is empty")
		case build.Arch == "":
			return errors.New("config: 'build[" + string(idx) + "] > arch' is empty")
		case build.Os == "":
			return errors.New("config: 'build[" + string(idx) + "] > os' is empty")
		}
	}
	return nil
}

func parseConfig(file string) (*AppConfig, error) {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var config AppConfig
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return nil, err
	}
	err = config.check()
	if err != nil {
		return nil, err
	}
	// set default values
	for _, build := range config.Build {
		if build.Dockerfile == "" {
			build.Dockerfile = "Dockerfile_" + build.Branch
		}
	}
	return &config, nil
}
