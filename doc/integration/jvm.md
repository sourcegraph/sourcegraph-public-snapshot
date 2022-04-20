# JVM dependencies integration with Sourcegraph

You can use Sourcegraph with JVM dependencies from any Maven repository, including Maven Central, Sonatype, or Artifactory.
This integration makes it possible to search and navigate through the source code of the JDK (Java standard library) or published Java, Scala, and Kotlin libraries (for example, [`com.google.guava:guava:27.0-jre`](https://sourcegraph.com/maven/com.google.guava/guava@v27.0-jre/-/blob/com/google/common/util/concurrent/Futures.java?L35)).

Feature | Supported?
------- | ----------
[Repository syncing](#repository-syncing) | ✅
[Credentials](#credentials) | ✅
[Rate limiting](#rate-limiting) | ✅
[Repository permissions](#repository-syncing) | ❌
[Multiple JVM dependencies code hosts](#multiple-jvm-dependency-code-hosts) | ❌

## Repository syncing

There are two ways to sync JVM dependency repositories.

* **LSIF** (recommended): run [`lsif-java`](https://sourcegraph.github.io/lsif-java/) against your JVM codebase and upload the generated index to Sourcegraph using the  [src-cli](https://github.com/sourcegraph/src-cli) command `src lsif upload`. Sourcegraph automatically synchronizes JVM dependency repositories based on the dependencies that are discovered by `lsif-java`.
* **Code host configuration**: manually list dependencies in the `"dependencies"` section of the JSON configuration when creating the JVM dependency code host. This method can be useful to verify that the credentials are picked up correctly without having to upload LSIF.

## Credentials

Sourcegraph uses [Coursier](https://get-coursier.io/) to resolve JVM dependencies.
Use the `"credentials"` section of the JSON configuration to provide usernames and passwords to access your Maven repository. See the Coursier documentation about [inline credentials](https://get-coursier.io/docs/other-credentials#inline) to learn more about how to format the `"credentials"` configuration.

## Rate limiting

By default, requests to the JVM dependency code host will be rate-limited
based on a default internal limit. ([source](https://github.com/sourcegraph/sourcegraph/blob/main/schema/jvm-packages.schema.json))

To manually set the value, add the following to your code host configuration:

```json
"rateLimit": {
  "enabled": true,
  "requestsPerHour": 600.0
}
```
where the `requestsPerHour` field is set based on your requirements.

**Not recommended**: Rate-limiting can be turned off entirely as well.
This increases the risk of overloading the code host.

```json
"rateLimit": {
  "enabled": false
}
```

## Repository permissions

⚠️ JVM dependency repositories are visible by all users of the Sourcegraph instance.

## Multiple JVM dependencies code hosts

⚠️ It's only possible to create one JVM dependency code host for each Sourcegraph instance.
See the issue [sourcegraph#32461](https://github.com/sourcegraph/sourcegraph/issues/32461) for more details about this limitation. In most situations, it's possible to work around this limitation by configurating multiple Maven repositories to the same JVM dependency code host.
