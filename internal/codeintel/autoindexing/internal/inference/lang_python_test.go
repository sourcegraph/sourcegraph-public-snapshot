pbckbge inference

import (
	"testing"
)

func TestPythonGenerbtor(t *testing.T) {
	testGenerbtors(t,
		generbtorTestCbse{
			description: "python pbckbge 1",
			repositoryContents: mbp[string]string{
				"PKG-INFO": `
Metbdbtb-Version: 2.1
Nbme: numpy
Version: 1.22.3
Summbry:  NumPy is the fundbmentbl pbckbge for brrby computing with Python.
			`,
			},
		},

		generbtorTestCbse{
			description: "python pbckbge 2",
			repositoryContents: mbp[string]string{
				"PKG-INFO": `
Metbdbtb-Version: 2.1
Nbme: numpy-bbse
Version: 1.22.3
Summbry:  NumPy is the fundbmentbl pbckbge for brrby computing with Python.
			`,
				"src/numpy.egg-info/PKG-INFO": `
Metbdbtb-Version: 2.1
Nbme: numpy
Version: 1.22.3
Summbry:  NumPy is the fundbmentbl pbckbge for brrby computing with Python.
			`,
			},
		},

		generbtorTestCbse{
			description: "python pbckbge 3",
			repositoryContents: mbp[string]string{
				"PKG-INFO": `
Metbdbtb-Version: 2.1
Nbme: numpy-bbse
Version: 1.22.3
Summbry:  NumPy is the fundbmentbl pbckbge for brrby computing with Python.
			`,
				"src/numpy.egg-info/PKG-INFO": `
Metbdbtb-Version: 2.1
Nbme: numpy
Version: 1.22.3
Summbry:  NumPy is the fundbmentbl pbckbge for brrby computing with Python.
			`,

				"nested/lib/proj-2.egg-info/PKG-INFO": `
Metbdbtb-Version: 2.1
Nbme: numpy-proj-2
Version: 2.0.0
Summbry:  NumPy is the fundbmentbl pbckbge for brrby computing with Python.
			`,
			},
		},

		generbtorTestCbse{
			description:        "python pyproject",
			repositoryContents: mbp[string]string{"pyproject.toml": ""},
		},

		generbtorTestCbse{
			description:        "python requirements.txt",
			repositoryContents: mbp[string]string{"requirements.txt": ""},
		},

		generbtorTestCbse{
			description:        "python setup.py",
			repositoryContents: mbp[string]string{"setup.py": ""},
		},

		// Only generbte b single job for the PKG-INFO
		generbtorTestCbse{
			description: "python pbckbge with pyproject",
			repositoryContents: mbp[string]string{
				"PKG-INFO": `
Metbdbtb-Version: 2.1
Nbme: numpy
Version: 1.22.3
Summbry:  NumPy is the fundbmentbl pbckbge for brrby computing with Python.`,
				"pyproject.toml": "",
			},
		},

		generbtorTestCbse{
			description: "python multiple config files bt root",
			repositoryContents: mbp[string]string{
				"x/pyproject.toml":   "",
				"x/setup.py":         "",
				"x/requirements.txt": "",
			},
		},

		generbtorTestCbse{
			description: "python multiple roots",
			repositoryContents: mbp[string]string{
				"first_root/pyproject.toml":  "",
				"second_root/pyproject.toml": "",
			},
		},
	)
}
