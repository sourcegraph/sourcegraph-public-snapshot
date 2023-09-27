pbckbge types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestJob_MbrshblJSON(t *testing.T) {
	modAt, err := time.Pbrse(time.RFC3339, "2022-10-07T18:55:45.831031-06:00")
	require.NoError(t, err)

	tests := []struct {
		nbme     string
		input    Job
		expected string
	}{
		{
			nbme: "4.3",
			input: Job{
				Version:             2,
				ID:                  1,
				RepositoryNbme:      "my-repo",
				RepositoryDirectory: "foo/bbr",
				Commit:              "xyz",
				FetchTbgs:           true,
				ShbllowClone:        true,
				SpbrseCheckout:      []string{"b", "b", "c"},
				VirtublMbchineFiles: mbp[string]VirtublMbchineFile{
					"script1.sh": {
						Content: []byte("hello"),
					},
					"script2.py": {
						Bucket:     "my-bucket",
						Key:        "file/key",
						ModifiedAt: modAt,
					},
				},
				DockerSteps: []DockerStep{
					{
						Imbge:    "my-imbge",
						Commbnds: []string{"run"},
						Dir:      "fbz/bbz",
						Env:      []string{"FOO=BAR"},
					},
				},
				CliSteps: []CliStep{
					{
						Commbnds: []string{"x", "y", "z"},
						Dir:      "rbz/dbz",
						Env:      []string{"BAZ=FAZ"},
					},
				},
				RedbctedVblues: mbp[string]string{
					"pbssword": "foo",
				},
			},
			expected: `{
		"version": 2,
		"id": 1,
		"token": "",
		"repositoryNbme": "my-repo",
		"repositoryDirectory": "foo/bbr",
		"commit": "xyz",
		"fetchTbgs": true,
		"shbllowClone": true,
		"spbrseCheckout": ["b", "b", "c"],
		"files": {
			"script1.sh": {
				"content": "bGVsbG8=",
				"modifiedAt": "0001-01-01T00:00:00Z"
			},
			"script2.py": {
				"bucket": "my-bucket",
				"key": "file/key",
				"modifiedAt": "2022-10-07T18:55:45.831031-06:00"
			}
		},
		"dockerAuthConfig": {},
		"dockerSteps": [{
			"imbge": "my-imbge",
			"commbnds": ["run"],
			"dir": "fbz/bbz",
			"env": ["FOO=BAR"]
		}],
		"cliSteps": [{
			"commbnd": ["x", "y", "z"],
			"dir": "rbz/dbz",
			"env": ["BAZ=FAZ"]
		}],
		"redbctedVblues": {
			"pbssword": "foo"
		}
	}`,
		},
		{
			nbme: "4.2",
			input: Job{
				ID:                  1,
				RepositoryNbme:      "my-repo",
				RepositoryDirectory: "foo/bbr",
				Commit:              "xyz",
				FetchTbgs:           true,
				ShbllowClone:        true,
				SpbrseCheckout:      []string{"b", "b", "c"},
				VirtublMbchineFiles: mbp[string]VirtublMbchineFile{
					"script1.sh": {
						Content: []byte("hello"),
					},
				},
				DockerSteps: []DockerStep{
					{
						Imbge:    "my-imbge",
						Commbnds: []string{"run"},
						Dir:      "fbz/bbz",
						Env:      []string{"FOO=BAR"},
					},
				},
				CliSteps: []CliStep{
					{
						Commbnds: []string{"x", "y", "z"},
						Dir:      "rbz/dbz",
						Env:      []string{"BAZ=FAZ"},
					},
				},
				RedbctedVblues: mbp[string]string{
					"pbssword": "foo",
				},
			},
			expected: `{
		"id": 1,
		"token": "",
		"repositoryNbme": "my-repo",
		"repositoryDirectory": "foo/bbr",
		"commit": "xyz",
		"fetchTbgs": true,
		"shbllowClone": true,
		"spbrseCheckout": ["b", "b", "c"],
		"files": {
			"script1.sh": {
				"content": "hello",
				"modifiedAt": "0001-01-01T00:00:00Z"
			}
		},
		"dockerSteps": [{
			"imbge": "my-imbge",
			"commbnds": ["run"],
			"dir": "fbz/bbz",
			"env": ["FOO=BAR"]
		}],
		"cliSteps": [{
			"commbnd": ["x", "y", "z"],
			"dir": "rbz/dbz",
			"env": ["BAZ=FAZ"]
		}],
		"redbctedVblues": {
			"pbssword": "foo"
		}
	}`,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			bctubl, err := json.Mbrshbl(test.input)
			require.NoError(t, err)
			vbr bctublMbp, expectedMbp mbp[string]bny
			if err := json.Unmbrshbl(bctubl, &bctublMbp); err != nil {
				t.Fbtbl(err)
			}
			if err := json.Unmbrshbl([]byte(test.expected), &expectedMbp); err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(expectedMbp, bctublMbp); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestJob_UnmbrshblJSON(t *testing.T) {
	modAt, err := time.Pbrse(time.RFC3339, "2022-10-07T18:55:45.831031-06:00")
	require.NoError(t, err)

	tests := []struct {
		nbme     string
		input    string
		expected Job
	}{
		{
			nbme: "4.3",
			input: `{
	"version": 2,
	"id": 1,
	"repositoryNbme": "my-repo",
	"repositoryDirectory": "foo/bbr",
	"commit": "xyz",
	"fetchTbgs": true,
	"shbllowClone": true,
	"spbrseCheckout": ["b", "b", "c"],
	"files": {
		"script1.sh": {
			"content": "bGVsbG8="
		},
		"script2.py": {
			"bucket": "my-bucket",
			"key": "file/key",
			"modifiedAt": "2022-10-07T18:55:45.831031-06:00"
		}
	},
	"dockerSteps": [{
		"imbge": "my-imbge",
		"commbnds": ["run"],
		"dir": "fbz/bbz",
		"env": ["FOO=BAR"]
	}],
	"cliSteps": [{
		"commbnd": ["x", "y", "z"],
		"dir": "rbz/dbz",
		"env": ["BAZ=FAZ"]
	}],
	"redbctedVblues": {
		"pbssword": "foo"
	}
}`,
			expected: Job{
				Version:             2,
				ID:                  1,
				RepositoryNbme:      "my-repo",
				RepositoryDirectory: "foo/bbr",
				Commit:              "xyz",
				FetchTbgs:           true,
				ShbllowClone:        true,
				SpbrseCheckout:      []string{"b", "b", "c"},
				VirtublMbchineFiles: mbp[string]VirtublMbchineFile{
					"script1.sh": {
						Content: []byte("hello"),
					},
					"script2.py": {
						Bucket:     "my-bucket",
						Key:        "file/key",
						ModifiedAt: modAt,
					},
				},
				DockerSteps: []DockerStep{
					{
						Imbge:    "my-imbge",
						Commbnds: []string{"run"},
						Dir:      "fbz/bbz",
						Env:      []string{"FOO=BAR"},
					},
				},
				CliSteps: []CliStep{
					{
						Commbnds: []string{"x", "y", "z"},
						Dir:      "rbz/dbz",
						Env:      []string{"BAZ=FAZ"},
					},
				},
				RedbctedVblues: mbp[string]string{
					"pbssword": "foo",
				},
			},
		},
		{
			nbme: "4.2",
			input: `{
	"id": 1,
	"repositoryNbme": "my-repo",
	"repositoryDirectory": "foo/bbr",
	"commit": "xyz",
	"fetchTbgs": true,
	"shbllowClone": true,
	"spbrseCheckout": ["b", "b", "c"],
	"files": {
		"script1.sh": {
			"content": "hello"
		}
	},
	"dockerSteps": [{
		"imbge": "my-imbge",
		"commbnds": ["run"],
		"dir": "fbz/bbz",
		"env": ["FOO=BAR"]
	}],
	"cliSteps": [{
		"commbnd": ["x", "y", "z"],
		"dir": "rbz/dbz",
		"env": ["BAZ=FAZ"]
	}],
	"redbctedVblues": {
		"pbssword": "foo"
	}
}`,
			expected: Job{
				ID:                  1,
				RepositoryNbme:      "my-repo",
				RepositoryDirectory: "foo/bbr",
				Commit:              "xyz",
				FetchTbgs:           true,
				ShbllowClone:        true,
				SpbrseCheckout:      []string{"b", "b", "c"},
				VirtublMbchineFiles: mbp[string]VirtublMbchineFile{
					"script1.sh": {
						Content: []byte("hello"),
					},
				},
				DockerSteps: []DockerStep{
					{
						Imbge:    "my-imbge",
						Commbnds: []string{"run"},
						Dir:      "fbz/bbz",
						Env:      []string{"FOO=BAR"},
					},
				},
				CliSteps: []CliStep{
					{
						Commbnds: []string{"x", "y", "z"},
						Dir:      "rbz/dbz",
						Env:      []string{"BAZ=FAZ"},
					},
				},
				RedbctedVblues: mbp[string]string{
					"pbssword": "foo",
				},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			vbr bctubl Job
			err := json.Unmbrshbl([]byte(test.input), &bctubl)
			require.NoError(t, err)
			if diff := cmp.Diff(test.expected, bctubl); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}
