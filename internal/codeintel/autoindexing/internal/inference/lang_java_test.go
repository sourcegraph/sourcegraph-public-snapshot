package inference

import (
	"testing"
)

func TestJavaGenerator(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
			description: "JVM project with lsif-java.json",
			repositoryContents: map[string]string{
				"lsif-java.json": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
		},
		generatorTestCase{
			description: "JVM project with Gradle",
			repositoryContents: map[string]string{
				"build.gradle": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
		},
		generatorTestCase{
			description: "JVM project with SBT",
			repositoryContents: map[string]string{
				"build.sbt": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
		},
		generatorTestCase{
			description: "JVM project with Maven",
			repositoryContents: map[string]string{
				"pom.xml": "",
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
		},
		generatorTestCase{
			description: "JVM project without build file",
			repositoryContents: map[string]string{
				"src/java/com/sourcegraph/codeintel/dumb.java": "",
				"src/java/com/sourcegraph/codeintel/fun.scala": "",
			},
		},
		generatorTestCase{
			description: "JVM project with Maven build file but no sources",
			repositoryContents: map[string]string{
				"pom.xml": "",
			},
		},
		generatorTestCase{
			description: "JVM project with Gradle build file but no sources",
			repositoryContents: map[string]string{
				"build.gradle": "",
			},
		},
		generatorTestCase{
			description: "JVM project with SBT build file but no sources",
			repositoryContents: map[string]string{
				"build.sbt": "",
			},
		},
		generatorTestCase{
			description: "JVM project with Mill build file but no sources",
			repositoryContents: map[string]string{
				"build.sc": "",
			},
		},
		generatorTestCase{
			description: "Nested JVM project with top-level build file",
			repositoryContents: map[string]string{
				"build.sbt": "",
				"my-module/src/java/com/sourcegraph/codeintel/dumb.java": "",
				"my-module/pom.xml": "",
			},
		},
		generatorTestCase{
			description: "Nested JVM project WITHOUT top-level build file",
			repositoryContents: map[string]string{
				"my-module/src/java/com/sourcegraph/codeintel/dumb.java": "",
				"my-module/pom.xml": "",
				"our-module/src/java/com/sourcegraph/codeintel/dumb.java": "",
				"our-module/pom.xml": "",
			},
		},
	)
}
