Run types are defined as the `RunType` enum type, and are used to indicate the type of pipeline that should be generated. What we run on pull requests, for example, is quite different from what gets run in our `main` branch, in order to keep the feedback loop short on pull requests.

> This notebook is part of the [Sourcegraph CI pipeline generator](https://sourcegraph.com/notebooks/Tm90ZWJvb2s6MTE3) notebook series.

```sourcegraph
repo:^github\.com/sourcegraph/sourcegraph$ file:dev/ci const (...RunType...) patterntype:structural count:all
```

Each `RunType` declares the conditions under which it should be used through a `RunTypeMatcher` to be returned in `(RunType).Matcher()`:

```sourcegraph
repo:^github\.com/sourcegraph/sourcegraph$ file:dev/ci func (...RunType) Matcher() ... {...} patterntype:structural count:1000
```

The computed `RunType` for a build can then be used as part of the pipeline generation process:

```sourcegraph
repo:^github\.com/sourcegraph/sourcegraph$ file:dev/ci ( case runtype.:[_]: OR .Is(...) ) patterntype:structural count:all -file:_test
```
