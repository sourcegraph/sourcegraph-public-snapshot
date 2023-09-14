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
	)
}
