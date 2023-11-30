package inference

import (
	"testing"
)

func TestPythonGenerator(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
			description: "python package 1",
			repositoryContents: map[string]string{
				"PKG-INFO": `
Metadata-Version: 2.1
Name: numpy
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.
			`,
			},
		},

		generatorTestCase{
			description: "python package 2",
			repositoryContents: map[string]string{
				"PKG-INFO": `
Metadata-Version: 2.1
Name: numpy-base
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.
			`,
				"src/numpy.egg-info/PKG-INFO": `
Metadata-Version: 2.1
Name: numpy
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.
			`,
			},
		},

		generatorTestCase{
			description: "python package 3",
			repositoryContents: map[string]string{
				"PKG-INFO": `
Metadata-Version: 2.1
Name: numpy-base
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.
			`,
				"src/numpy.egg-info/PKG-INFO": `
Metadata-Version: 2.1
Name: numpy
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.
			`,

				"nested/lib/proj-2.egg-info/PKG-INFO": `
Metadata-Version: 2.1
Name: numpy-proj-2
Version: 2.0.0
Summary:  NumPy is the fundamental package for array computing with Python.
			`,
			},
		},

		generatorTestCase{
			description:        "python pyproject",
			repositoryContents: map[string]string{"pyproject.toml": ""},
		},

		generatorTestCase{
			description:        "python requirements.txt",
			repositoryContents: map[string]string{"requirements.txt": ""},
		},

		generatorTestCase{
			description:        "python setup.py",
			repositoryContents: map[string]string{"setup.py": ""},
		},

		// Only generate a single job for the PKG-INFO
		generatorTestCase{
			description: "python package with pyproject",
			repositoryContents: map[string]string{
				"PKG-INFO": `
Metadata-Version: 2.1
Name: numpy
Version: 1.22.3
Summary:  NumPy is the fundamental package for array computing with Python.`,
				"pyproject.toml": "",
			},
		},

		generatorTestCase{
			description: "python multiple config files at root",
			repositoryContents: map[string]string{
				"x/pyproject.toml":   "",
				"x/setup.py":         "",
				"x/requirements.txt": "",
			},
		},

		generatorTestCase{
			description: "python multiple roots",
			repositoryContents: map[string]string{
				"first_root/pyproject.toml":  "",
				"second_root/pyproject.toml": "",
			},
		},
	)
}
