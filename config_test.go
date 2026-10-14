package josuke

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_I_Can_Parse_YAML(t *testing.T) {
	trial1, err := parseConfig("testdata/test.config.yaml")
	envFile, absErr := filepath.Abs("testdata/test.env")
	require.NoError(t, absErr)
	goal1 := Josuke{
		Hooks: []*Hook{{
			Name:   "github",
			Type:   "github",
			Path:   "/josuke/github",
			Secret: "cabane123",
		}},
		Port:             8082,
		LogLevel:         0,
		LogLevelName:     "INFO",
		Host:             "localhost",
		Cert:             "",
		Key:              "",
		Store:            "",
		HealthcheckRoute: "",
		Deployment: []*Repo{
			{
				Name:         "monkeydioude/josuke",
				EnvFile:      envFile,
				Environments: []string{"STORAGE_DIR=/path/to/some/dir"},
				Branches: []Branch{
					{
						Name:         "master",
						Environments: []string{"STORAGE_DIR=/path/to/some/different/dir"},
						Actions: []Action{
							{
								Action:       "push",
								Environments: []string{"STORAGE_DIR=/path/to/some/different/dir/once_more"},
								Commands: [][]string{
									{"cd", "%base_dir%"},
									{"git", "clone", "git@github.com:monkeycddioude/josuke.git"},
									{"cd", "%proj_dir%"},
									{"git", "checkout", "master"},
									{"git", "fetch", "--all"},
									{"git", "reset", "--hard", "origin/master"},
									{"cd", "bin/josuke"},
									{"go", "install"},
									{"service", "josuke", "restart"},
								},
							},
						},
					},
				},
				BaseDir: "/test/github.com/monkeydioude",
				ProjDir: "cabane",
			},
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, goal1, *trial1)
}

func writeConfig(t *testing.T, dir string, content string) string {
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func Test_Config_With_A_Missing_Env_File_Does_Not_Start(t *testing.T) {
	path := writeConfig(t, t.TempDir(), `
deployment:
  - repo: "monkeydioude/josuke"
    env_file: "missing.env"
`)

	_, err := New(path)

	assert.ErrorContains(t, err, "repo monkeydioude/josuke: could not read env file")
	assert.ErrorContains(t, err, "missing.env")
}

func Test_Config_With_An_Invalid_Env_File_Does_Not_Start(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), []byte("A=1\nnot a variable secret\n"), 0o644))
	path := writeConfig(t, dir, `
deployment:
  - repo: "monkeydioude/josuke"
    env_file: ".env"
`)

	_, err := New(path)

	assert.ErrorContains(t, err, "line 2: expected KEY=VALUE")
	assert.NotContains(t, err.Error(), "secret")
}

func Test_Config_With_An_Invalid_Environment_Does_Not_Start(t *testing.T) {
	path := writeConfig(t, t.TempDir(), `
deployment:
  - repo: "monkeydioude/josuke"
    branches:
      - branch: "master"
        actions:
          - action: "push"
            environments:
              - STORAGE_DIR
`)

	_, err := New(path)

	assert.EqualError(t, err, `invalid deployment environment: repo monkeydioude/josuke, branch master, action push: invalid environment "STORAGE_DIR", expected KEY=VALUE`)
}

func Test_Boot_Logs_The_Env_File_And_The_Variable_Keys_Only(t *testing.T) {
	var buf bytes.Buffer
	defer log.SetOutput(log.Writer())
	log.SetOutput(&buf)
	envFile, err := filepath.Abs("testdata/test.env")
	require.NoError(t, err)

	_, err = New("testdata/test.config.yaml")

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "[INFO] monkeydioude/josuke env file "+envFile+": STORAGE_DIR, DB_URL, LOVE_ME_LOVE_ME\n")
	assert.Contains(t, buf.String(), "[INFO] monkeydioude/josuke environments: STORAGE_DIR\n")
	assert.Contains(t, buf.String(), "[INFO] monkeydioude/josuke@master environments: STORAGE_DIR\n")
	assert.Contains(t, buf.String(), "[INFO] monkeydioude/josuke@master push environments: STORAGE_DIR\n")
	assert.NotContains(t, buf.String(), "/path/")
	assert.NotContains(t, buf.String(), "postgres")
	assert.NotContains(t, buf.String(), "say that you love me")
}
