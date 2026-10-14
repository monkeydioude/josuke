package josuke

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
)

// extraEnv returns the variables the commands get on top of josuke's environment:
// the env file ones, overridden by the environments of the config.
// A key can be repeated, exec.Cmd keeps its last value.
func (i *Info) extraEnv(logger *log.Logger) ([]string, error) {
	var env []string
	if i.EnvFile != "" {
		fileEnv, err := readEnvFile(i.EnvFile)
		if err != nil {
			return nil, err
		}
		logEnvKeys(logger, "env file "+i.EnvFile, fileEnv)
		env = fileEnv
	}
	if len(i.Env) > 0 {
		logEnvKeys(logger, "environments", i.Env)
	}
	return append(env, i.Env...), nil
}

// logEnvKeys logs the keys of the variables, never their values: they could be secrets.
func logEnvKeys(logger *log.Logger, source string, env []string) {
	keys := "(empty)"
	if len(env) > 0 {
		keys = strings.Join(envKeys(env), ", ")
	}
	logger.Printf("[INFO] %s: %s\n", source, keys)
}

// envKeys returns the keys of the variables, without duplicates.
func envKeys(env []string) []string {
	keys := make([]string, 0, len(env))
	for _, v := range env {
		key, _, _ := strings.Cut(v, "=")
		if !slices.Contains(keys, key) {
			keys = append(keys, key)
		}
	}
	return keys
}

// readEnvFile parses KEY=VALUE lines, ignoring blank lines and # comments.
// An "export " prefix and quotes around the value are removed. Variables are not expanded.
func readEnvFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read env file: %w", err)
	}
	var env []string
	for n, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(strings.TrimPrefix(line, "export "), "=")
		key = strings.TrimSpace(key)
		// The line is not quoted in the error, the value could be a secret.
		if !ok || !validEnvKey(key) {
			return nil, fmt.Errorf("env file %s, line %d: expected KEY=VALUE", path, n+1)
		}
		env = append(env, key+"="+unquote(strings.TrimSpace(value)))
	}
	return env, nil
}

func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' || s[0] == '\'') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}

// checkEnvironments fails if a variable is not KEY=VALUE.
func checkEnvironments(env []string) error {
	for _, v := range env {
		key, _, ok := strings.Cut(v, "=")
		// Only the key is quoted in the error, the value could be a secret.
		if !ok || !validEnvKey(key) {
			return fmt.Errorf("invalid environment %q, expected KEY=VALUE", key)
		}
	}
	return nil
}

func validEnvKey(key string) bool {
	return key != "" && !strings.ContainsAny(key, " \t")
}
