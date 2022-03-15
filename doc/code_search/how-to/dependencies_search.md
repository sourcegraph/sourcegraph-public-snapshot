# Dependencies search <span class="badge badge-beta">beta</span>

Dependencies search is a code search feature that lets you search through the dependencies of your repositories.


<img src="https://storage.googleapis.com/sourcegraph-assets/docs/images/code_search/dependencies-search-usage.png" style="margin-left:0;margin-right:0;"/>

### Setup

Configure a package host connection for each kind of dependency you want to search over. Currently only [npm dependencies](../../integration/npm.md) are supported.

### Use cases

Resolve an incident faster by [quickly finding where an error comes from](https://sourcegraph.com/search?q=context:global+repo:deps%28%5Egithub%5C.com/sourcegraph/sourcegraph%24%403.37%29+Observable+cannot+be+called+as+a+function&patternType=literal) and then understanding the code around it by leveraging [code intelligence](../../code_intelligence/explanations/features.md).<br/>

```sgquery
r:deps(^github\.com/sourcegraph/sourcegraph$@3.37) Observable cannot be called as a function
```

### Compatibility

The following table outlines the kinds of dependency repositories that dependency search supports how it finds those dependencies in your repositories.

Kind | How | Supported
---- | ------ | ---------
[npm](../../integration/npm.md) | `package-lock.json` | ✅
[npm](../../integration/npm.md) | `yarn.lock` | ✅
[JVM](../../integration/jvm.md) | `gradle.lockfile` | ❌
[JVM](../../integration/jvm.md) | `pom.xml` | ❌
Go | `go.sum` | ❌
Python | `poetry.lock` | ❌

### Reference

- [`repo:dependencies(...)`](../reference/language.md#repo-dependencies)
