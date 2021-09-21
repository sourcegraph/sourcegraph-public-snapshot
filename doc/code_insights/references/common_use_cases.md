# Common Code Insights use cases and recipes

Here are some common use cases for Code Insights and example data series queries you could use. 

For all use cases, you can also explore your insight by [filtering repositories in real time](../how-tos/filtering_an_insight.md) or add any [Sourcegraph search filter](../../../code_search/reference/language.md#search-pattern) to the data series query to filter by language, directory, or content. Currently, the sample queries using commit and diff searches are only supported for insights running over explicit lists of specific repositories. 

*The sample queries below make the assumption you [do not want to search fork or archived](../references/common_reasons_code_insights_may_not_match_search_results.md#not-including-fork-no-and-archived-no-in-your-insight-query) repositories. You can exclude those flags if you do.*

## Migration tracking 

**How many repos yet contain a config or docs file in a specific directory?**
```sgquery
select:repo file:docs/*/new_config_file archived:no fork:no
```

**How is the switchover from files containing “blacklist/whitelist” to “denylist/allowlist" going?**
```sgquery
// series 1, decreasing
select:file blacklist OR whitelist archived:no fork:no 
// series 2, increasing
select:file denylist OR allowlist archived:no fork:no 
```

**Tracking migration from global CSS to CSS modules**
```sgquery
// series 1, decreasing
type:file lang:scss -file:module.scss patterntype:regexp archived:no fork:no 
// series 2, increasing
type:file lang:scss file:module.scss patterntype:regexp archived:no fork:no 
```
Note these queries differ only by `-file:` exclusion vs `file:` inclusion – because you can know that all `lang:scss` files are either modules or not.


## Adoption tracking

**How many repos or teams are using a new API your team built?**
```sgquery
select:repo ourApiLibrary.load archived:no fork:no
```

**Are more organizations using yarn and should we support that at a company level?**
```sgquery
select:repo file:yarn.lock archived:no fork:no
```

**Which databases are we calling or writing to most often?**
```sgquery
// redis
redis\.set\(.*\) patternType=regexp archived:no fork:no 
// graphQL
graphql\(.*\) patternType=regexp archived:no fork:no
```

**Are a growing number of repos importing a large/expensive package?**
```sgquery
select:repo import\slargePkg\s patternType=regexp archived:no fork:no
```

## Deprecation tracking

**Have we removed all instances of a deprecated CSS class or icon?**
```sgquery
2018logo.png archived:no fork:no
```
```sgquery
theme-redesign archived:no fork:no
```

**How are we progressing on deprecating tooling that we’re moving off of?**
```sgquery
slowEventLib.log archived:no fork:no
```
Or you can count how many removals (below) rather than how many remain (above): 
```sgquery
slowEventLib.log type:diff select:commit.diff.removed archived:no fork:no
```

**Are we deprecating a structural code pattern in favor of a safer pattern?** 

Example: do all our tries have catches? This tracks how many do not: 
```sgquery
try {:[_]} catch (:[e]) { } finally {:[_]} lang:java patternType:structural archived:no fork:no
```

## Code hygiene and health 

**How many TODOs do we have in a specific part of the codebase (or all of it)?** 
```sgquery
TODO archived:no fork:no
```

**How many linter override rules do we have?**
```sgquery
file:^\.eslintignore .\n patternType:regexp archived:no fork:no
```
(This counts the number of lines, which are file paths to ignore, in .eslintignore files)

**How frequently do we have commits with “revert” in the commit message?**
```sgquery
type:commit revert archived:no fork:no
```

**How many times are we using deprecated calls?**
```sgquery
lang:java @deprecated archived:no fork:no
```

**How many repos have CODEOWNERS files?** 
```sgquery
\\ how many do:
file:CODEOWNERS select:repo archived:no fork:no
\\ how many don't:
-file:CODEOWNERS select:repo archived:no fork:no
```

## Security vulnerabilities

**Has a vulnerable open source library been fully removed, or how quickly is the deprecation progressing?**
```sgquery
vulnerableLib@14.3.9 archived:no fork:no
```

**How quickly do we notice and remove API keys when they are committed?** 
```sgquery
regexMatchingAPIKey patternType:regexp archived:no fork:no
```

**How often are we merging permissions changes?**
```sgquery
type:commit perms|permissions patternType:regexp archived:no fork:no
```

## Version tracking (packages or languages)

**Which package version are parts of the codebase using; what should we align on?**
```sgquery
// for version 13
nvm install 13 archived:no fork:no
// for version 14
nvm install 14 archived:no fork:no
// ... so on
```

**Which language versions are in use most? How are we tracking on migrating them?**  
```sgquery
// python 2.7
#!/usr/bin/env python2.7 archived:no fork:no
// python 3
#!/usr/bin/env python3 archived:no fork:no
```

## Codebase Topline Metrics
> Note that some of these may be "extreme" and perform slower than an average insight. 

**What is our codebase size in LOC (and is it growing/shrinking)?** 
```sgquery 
.\n patternType:regex archived:no fork:no
```

**What is our codebase size for a specific language?**
```sgquery 
.\n lang:TypeScript patternType:regex archived:no fork:no
```

**Do we have more Typescript or more Go files?** 
```sgquery 
select:file lang:TypeScript archived:no fork:no
// vs Go files
select:file lang:Go archived:no fork:no
```

**What number of iOS app screens are in the entire app?**
```sgquery
struct : view$ patternType:regexp lang:swift archived:no fork:no
```

## Understanding Code by Team

**Which teams or repos adopted our new events API fastest?** 
```sgquery
newEventAPI.call archived:no fork:no
```
And then [filter by repositories](../how-tos/filtering_an_insight.md) in real time. 

**Which teams have the most usage of an API we’re having problems with?**
```sgquery
// series 1
problemAPI file:teamOneFilePath archived:no fork:no
// series 2
problemAPI file:teamTwoFilePath archived:no fork:no
// ... so on
```
