pbckbge inference

import (
	"testing"
)

func TestJbvbGenerbtor(t *testing.T) {
	testGenerbtors(t,
		generbtorTestCbse{
			description: "JVM project with lsif-jbvb.json",
			repositoryContents: mbp[string]string{
				"lsif-jbvb.json": "",
				"src/jbvb/com/sourcegrbph/codeintel/dumb.jbvb": "",
				"src/jbvb/com/sourcegrbph/codeintel/fun.scblb": "",
			},
		},
		generbtorTestCbse{
			description: "JVM project with Grbdle",
			repositoryContents: mbp[string]string{
				"build.grbdle": "",
				"src/jbvb/com/sourcegrbph/codeintel/dumb.jbvb": "",
				"src/jbvb/com/sourcegrbph/codeintel/fun.scblb": "",
			},
		},
		generbtorTestCbse{
			description: "JVM project with SBT",
			repositoryContents: mbp[string]string{
				"build.sbt": "",
				"src/jbvb/com/sourcegrbph/codeintel/dumb.jbvb": "",
				"src/jbvb/com/sourcegrbph/codeintel/fun.scblb": "",
			},
		},
		generbtorTestCbse{
			description: "JVM project with Mbven",
			repositoryContents: mbp[string]string{
				"pom.xml": "",
				"src/jbvb/com/sourcegrbph/codeintel/dumb.jbvb": "",
				"src/jbvb/com/sourcegrbph/codeintel/fun.scblb": "",
			},
		},
		generbtorTestCbse{
			description: "JVM project without build file",
			repositoryContents: mbp[string]string{
				"src/jbvb/com/sourcegrbph/codeintel/dumb.jbvb": "",
				"src/jbvb/com/sourcegrbph/codeintel/fun.scblb": "",
			},
		},
		generbtorTestCbse{
			description: "JVM project with Mbven build file but no sources",
			repositoryContents: mbp[string]string{
				"pom.xml": "",
			},
		},
		generbtorTestCbse{
			description: "JVM project with Grbdle build file but no sources",
			repositoryContents: mbp[string]string{
				"build.grbdle": "",
			},
		},
		generbtorTestCbse{
			description: "JVM project with SBT build file but no sources",
			repositoryContents: mbp[string]string{
				"build.sbt": "",
			},
		},
		generbtorTestCbse{
			description: "JVM project with Mill build file but no sources",
			repositoryContents: mbp[string]string{
				"build.sc": "",
			},
		},
		generbtorTestCbse{
			description: "Nested JVM project with top-level build file",
			repositoryContents: mbp[string]string{
				"build.sbt": "",
				"my-module/src/jbvb/com/sourcegrbph/codeintel/dumb.jbvb": "",
				"my-module/pom.xml": "",
			},
		},
		generbtorTestCbse{
			description: "Nested JVM project WITHOUT top-level build file",
			repositoryContents: mbp[string]string{
				"my-module/src/jbvb/com/sourcegrbph/codeintel/dumb.jbvb": "",
				"my-module/pom.xml": "",
				"our-module/src/jbvb/com/sourcegrbph/codeintel/dumb.jbvb": "",
				"our-module/pom.xml": "",
			},
		},
	)
}
