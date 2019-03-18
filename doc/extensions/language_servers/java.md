# Java: code intelligence configuration

This page describes additional configuration that may be needed for Java code intelligence on certain codebases. To enable Java code intelligence, see the [installation documentation](install/index.md).

---

## Cross-repository code intelligence

Jump-to-definition and find-references actions in Sourcegraph are source-to-source. When a user performs a jump-to-def to a dependency or a find-references in external repositories, they will navigate to the original source (rather than a decompiled `.class` file). In order for this to work, the source of the dependency or dependee must be available on your Sourcegraph instance. Refer to the [instructions for adding repositories](../../../admin/repo/add.md).

In addition to your organization's repositories and your open-source dependencies, it is also common to add the following repositories (or forks thereof):

- Java JDK
  - https://github.com/jdkmirrors/openjdk
- Android
  - https://github.com/androidmirrors/android-sdk
  - [https://github.com/androidmirrors/platform_frameworks_testing](https://github.com/androidmirrors/platform_frameworks_testing)

---

## Custom configuration

If Java code intelligence does not work or only partially works, you most likely need to configure it to work with your build system. Refer to the following configuration options.

### Build configuration

Most Java projects built with Maven and Gradle are supported automatically. In cases where neither is used or where the build is sufficiently complex, add a `javaconfig.json` file to the root directory of the source repository. This file specifies the information Sourcegraph needs to provide code intelligence.

Here is an example `javaconfig.json` that could be constructed for the [Apache Commons IO](https://sourcegraph.com/github.com/apache/commons-io) project (if it didn't already include Maven build files):

```json
{
  "projects": {
    "/": {
      // Metadata about the artifact if this project can be depended on by other projects
      "artifactId": "commons-io",
      "groupId": "commons-io",
      "version": "2.6-SNAPSHOT",

      // The source directories of this project
      "sourceDirectories": ["src/main/java"],

      // The test source directories of this project
      "testSourceDirectories": ["src/test/java"],

      // The Maven repositories from which to fetch dependencies
      "repositories": [
        {
          "id": "apache.snapshots",
          "url": "https://repository.apache.org/snapshots"
        },
        {
          "id": "central",
          "url": "https://repo.maven.apache.org/maven2"
        }
      ],

      // Dependencies specified by their artifact metadata
      "dependencies": [
        {
          "groupId": "junit",
          "artifactId": "junit",
          "version": "4.12",
          "scope": "test"
        },
        {
          "groupId": "org.hamcrest",
          "artifactId": "hamcrest-core",
          "version": "1.3",
          "scope": "test"
        }
      ],

      // Additional options to pass to javac during compilation
      "compilerOptions": []
    }
  }
}
```

If you are running Gradle, you can use the [Javaconfig Gradle plugin](https://plugins.gradle.org/plugin/com.sourcegraph.gradle) to easily generate a `javaconfig.json` from your existing Gradle build files. Follow the instructions in the link to add the plugin. Then run `./gradlew javaconfig -P outputFile=javaconfig.json` and commit the generated `javaconfig.json` file.

### Private artifact repositories

If you have a private artifact repository that requires authentication, add the credentials to the
`initializationOptions.servers` field in the site config for Java:

```json
{
  ...
  "langservers": [
    ...
    {
      ...
      "language": "java",
      "initializationOptions": {
        "servers": [
          {
            "id": "${ARTIFACT_REPOSITORY_ID}",
            "username": "${ARTIFACT_REPOSITORY_USERNAME}",
            "password": "${ARTIFACT_REPOSITORY_PASSWORD"
          }
        ]
      }
    }
    ...
  ]
  ...
}
...
```

The values for the `id`, `username`, and `password` fields can be obtained by running the following
in your development environment:

```
mvn help:effective-settings -DshowPasswords=true
```

You can specify multiple artifact repositories in the `servers` field and should specify one entry
for each private artifact repository you have.

If the artifact repository is unauthenticated, then the above config is unnecessary.

It is recommended that you add the source repositories of all JARs hosted in your private artifact
repositories, so that users will be able to jump through to the source code of these dependencies.

### Maven plugins

Maven plugins will not be executed. This includes the `maven replacer` plugin, which is often used to generate Java source files from templates.

### Gradle execution

> WARNING: Security note: Before enabling this option, turn off automatic-repository cloning in your site config. After enabling this option, do not add any repository to Sourcegraph whose build scripts you do not trust. This option enables the Java language service to execute the Gradle build scripts of repositories added to your Sourcegraph instance.

Java code intelligence works automatically for most Gradle build scripts. If your Gradle build is more complex, add the following to your site config:

```json
  "executeGradleOriginalRootPaths": "git://${REPOSITORY_PATH_1}?%,git://${REPOSITORY_PATH_2}?%"
```

Replace `${REPOSITORY_PATH_1}` and `${REPOSITORY_PATH_2}` with repositories for which you want this option enabled. The repository path is the segment of the repository URL that identifies the repository (e.g., `github.com/mockito/mockito`).

---

## Configuring the JVM

Use the `JVM_OPT` environment variable to pass JVM options to the Java language server. This is sometimes necessary when analyzing large Java projects, which consumes a large amount of memory. For example:

```text
JVM_OPT='-Xms8000m -Xmx8000m -Dsun.zip.disableMemoryMapping=true'
```

- For single-node Sourcegraph deployments, set this environment variable on the container running the `sourcegraph/codeintel-java` image: `docker run ... -e JVM_OPT='...' sourcegraph/codeintel-java`
- For Sourcegraph cluster deployments on Kubernetes, set this environment variable on the `xlang-java` Kubernetes deployment. You will need to reapply this change it each time you `helm upgrade`; this is a known issue that will be addressed in a future release.

### Heap size

If you notice the Java language server exiting unexpectedly with an out-of-memory error or a `Killed` message, increase the JVM heap size (using `JVM_OPT` as shown above). The parameters `-Xmx8000m -Xms8000m` (8000 MB) are usually sufficient. You can monitor actual memory usage and reduce the heap size accordingly to preserve resources if needed.

As a rough guideline, the Java language server generally consumes about 50% as much memory as a local Java IDE (IntelliJ/Eclipse/etc.) analyzing the samee project.
