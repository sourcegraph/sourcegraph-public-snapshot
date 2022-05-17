# Dependencies search <span class="badge badge-beta">beta</span>

Dependencies search is a code search feature that lets you search through the dependencies of your repositories.


<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code_search/dependencies-search-usage.png" style="margin-left:0;margin-right:0;"/>

### Setup

Configure a package host connection for each kind of dependency you want to search over.

### Use cases

Resolve an incident faster by [quickly finding where an error comes from](https://sourcegraph.com/search?q=context:global+repo:deps%28%5Egithub%5C.com/sourcegraph/sourcegraph%24%403.37%29+Observable+cannot+be+called+as+a+function&patternType=literal) and then understanding the code around it by leveraging [code intelligence](../../code_intelligence/explanations/features.md).<br/>

```sgquery
r:deps(^github\.com/sourcegraph/sourcegraph$@3.37) Observable cannot be called as a function
```

Search only npm dependencies:

```sgquery
r:deps(^github\.com/sourcegraph/sourcegraph$@3.37) r:^npm throw
```

Search only Go dependencies:

```sgquery
r:deps(^github\.com/sourcegraph/sourcegraph$@3.37) r:^go fmt.Println
```

Search only Python dependencies:

```sgquery
r:deps(^github\.com/firecracker-microvm/firecracker$) r:^python
```

### Compatibility

The following table outlines the kinds of dependency repositories that dependency search supports and how it finds those dependencies in your repositories.

Kind                                  | How                       | Direct | Transitive
-------------------------------       |-------------------------- |------- | ----------
[npm](../../integration/npm.md)       | lsif-typescript uploads   | ✅     | ✅
[npm](../../integration/npm.md)       | `package-lock.json`       | ✅     | ✅
[npm](../../integration/npm.md)       | `yarn.lock`               | ✅     | ✅
[Python](../../integration/python.md) | lsif-python uploads       | ❌     | ❌
[Python](../../integration/python.md) | `poetry.lock`             | ✅     | ✅
[Python](../../integration/python.md) | `Pipfile.lock`            | ✅     | ✅
[Go](../../integration/go.md)         | lsif-go uploads           | ❌     | ❌
[Go](../../integration/go.md)         | `go.mod`                  | ✅     | ✅ with Go >= 1.17 go.mod files
[JVM](../../integration/jvm.md)       | lsif-java uploads         | ❌     | ❌
[JVM](../../integration/jvm.md)       | `gradle.lockfile`         | ❌     | ❌
[JVM](../../integration/jvm.md)       | `pom.xml`                 | ❌     | ❌

### Reference

- [`repo:dependencies(...)`](../reference/language.md#repo-dependencies)
