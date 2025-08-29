package josuke

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_I_Can_Parse_YAML(t *testing.T) {
	trial1, err := parseConfig("testdata/test.config.yaml")
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
				Name: "monkeydioude/josuke",
				Branches: []Branch{
					{
						Name: "master",
						Actions: []Action{
							{
								Action: "push",
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
