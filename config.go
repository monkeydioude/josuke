package josuke

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	yaml "github.com/goccy/go-yaml"
)

func parseConfig(configFilePath string) (*Josuke, error) {
	file, err := os.ReadFile(configFilePath)

	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	j := &Josuke{
		LogLevelName: "INFO",
		Host:         "localhost",
		Port:         8082,
		Hooks:        make([]*Hook, 0),
	}

	if err := json.Unmarshal(file, j); err != nil {
		if err := yaml.Unmarshal(file, j); err != nil {
			return nil, errors.New("could not parse yaml/json from config file")
		}
	}
	absConfigFilePath, err := filepath.Abs(configFilePath)
	if err != nil {
		return nil, err
	}
	// Relative to the config file rather than to the working directory: commands change it.
	for _, repo := range j.Deployment {
		if repo.EnvFile != "" && !filepath.IsAbs(repo.EnvFile) {
			repo.EnvFile = filepath.Join(filepath.Dir(absConfigFilePath), repo.EnvFile)
		}
	}
	return j, nil
}

// checkDeploymentEnv fails if an environment variable or an env file is invalid,
// so a typo is caught at boot rather than at the first job. It logs the keys of the variables.
func (j *Josuke) checkDeploymentEnv() error {
	logEnv := func(source string, env []string) {
		if j.LogEnabled(InfoLevel) && len(env) > 0 {
			logEnvKeys(log.Default(), source, env)
		}
	}
	for _, repo := range j.Deployment {
		if repo.EnvFile != "" {
			env, err := readEnvFile(repo.EnvFile)
			if err != nil {
				return fmt.Errorf("repo %s: %w", repo.Name, err)
			}
			if j.LogEnabled(InfoLevel) {
				logEnvKeys(log.Default(), repo.Name+" env file "+repo.EnvFile, env)
			}
		}
		if err := checkEnvironments(repo.Environments); err != nil {
			return fmt.Errorf("repo %s: %w", repo.Name, err)
		}
		logEnv(repo.Name+" environments", repo.Environments)
		for _, branch := range repo.Branches {
			if err := checkEnvironments(branch.Environments); err != nil {
				return fmt.Errorf("repo %s, branch %s: %w", repo.Name, branch.Name, err)
			}
			logEnv(repo.Name+"@"+branch.Name+" environments", branch.Environments)
			for _, action := range branch.Actions {
				if err := checkEnvironments(action.Environments); err != nil {
					return fmt.Errorf("repo %s, branch %s, action %s: %w", repo.Name, branch.Name, action.Action, err)
				}
				logEnv(repo.Name+"@"+branch.Name+" "+action.Action+" environments", action.Environments)
			}
		}
	}
	return nil
}
