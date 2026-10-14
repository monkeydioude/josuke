package josuke

import (
	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Read_Env_File(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "# comment\n\nA=1\r\nexport B = two words \nC=\"quoted # not a comment\"\nD='single'\nE=\nF=a=b\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

	env, err := readEnvFile(path)

	require.NoError(t, err)
	assert.Equal(t, []string{"A=1", "B=two words", "C=quoted # not a comment", "D=single", "E=", "F=a=b"}, env)
}

func Test_Commands_Get_The_Environments_From_The_Most_Specific_Level(t *testing.T) {
	skipOnWindows(t)
	t.Setenv("V0", "josuke")
	t.Setenv("V1", "josuke")
	envFile := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(envFile, []byte("V1=file\nV2=file\nV3=file\nV4=file\nV5=file\n"), 0o644))
	deployment := []*Repo{{
		Name:         "monkeydioude/josuke",
		EnvFile:      envFile,
		Environments: []string{"V2=repo", "V3=repo", "V4=repo", "V5=repo"},
		Branches: []Branch{{
			Name:         "master",
			Environments: []string{"V3=branch", "V4=branch", "V5=branch"},
			Actions: []Action{{
				Action:       "push",
				Environments: []string{"V4=action", "V5=action"},
				Commands: [][]string{
					{"sh", "-c", "echo first $V0 $V1 $V2 $V3 $V4 $V5"},
					{"sh", "-c", "echo second $V5"},
				},
			}},
		}},
	}}
	payload := &Payload{Ref: "refs/heads/master", Action: "push", Repository: Repository{Name: "monkeydioude/josuke"}}
	var buf bytes.Buffer

	action, info := payload.getDeployAction(deployment, "", "github", "push")
	require.NotNil(t, action)
	require.NoError(t, action.execute(context.Background(), info, log.New(&buf, "", 0)))

	assert.Contains(t, buf.String(), "[INFO] env file "+envFile+": V1, V2, V3, V4, V5\n")
	assert.Contains(t, buf.String(), "[INFO] environments: V2, V3, V4, V5\n")
	assert.Contains(t, buf.String(), "sh | first josuke file repo branch action action\n")
	assert.Contains(t, buf.String(), "sh | second action\n")
	// The config is not changed by the merge.
	assert.Equal(t, []string{"V2=repo", "V3=repo", "V4=repo", "V5=repo"}, deployment[0].Environments)
}

func Test_Env_File_Is_Read_At_Each_Run(t *testing.T) {
	skipOnWindows(t)
	envFile := filepath.Join(t.TempDir(), ".env")
	action := &Action{Action: "push", Commands: [][]string{{"sh", "-c", "echo $V"}}}
	info := &Info{EnvFile: envFile}
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	require.NoError(t, os.WriteFile(envFile, []byte("V=before\n"), 0o644))
	require.NoError(t, action.execute(context.Background(), info, logger))
	require.NoError(t, os.WriteFile(envFile, []byte("V=after\n"), 0o644))
	require.NoError(t, action.execute(context.Background(), info, logger))
	require.NoError(t, os.Remove(envFile))
	err := action.execute(context.Background(), info, logger)

	assert.Contains(t, buf.String(), "sh | before\n")
	assert.Contains(t, buf.String(), "sh | after\n")
	assert.ErrorContains(t, err, "could not read env file")
}
