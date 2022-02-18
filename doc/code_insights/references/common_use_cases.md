# Common Code Insights use cases and recipes

Here are some common use cases for Code Insights and example data series queries you could use. 

For all use cases, you can also explore your insight by [filtering repositories in real time](../how-tos/filtering_an_insight.md) or add any [Sourcegraph search filter](../../../code_search/reference/language.md#search-pattern) to the data series query to filter by language, directory, or content. Currently, the sample queries using commit and diff searches are only supported for insights running over explicit lists of specific repositories. 

*The sample queries below make the assumption you [do not want to search fork or archived](../references/common_reasons_code_insights_may_not_match_search_results.md#not-including-fork-no-and-archived-no-in-your-insight-query) repositories. You can exclude those flags if you do.*

<!-- everything below this line should be generated according to https://github.com/sourcegraph/sourcegraph/pull/31469 -->

## Popular


### Terraform versions
Detect and track which Terraform versions are present or most popular in your codebase
```sgquery
app.terraform.io/(.*)\n version =(.*)1.1.0 patternType:regexp lang:Terraform archived:no fork:no
```
```sgquery
app.terraform.io/(.*)\n version =(.*)1.2.0 patternType:regexp lang:Terraform archived:no fork:no
```


### Global CSS to CSS modules
Tracking migration from global CSS to CSS modules
```sgquery
select:file lang:SCSS -file:module patterntype:regexp archived:no fork:no
```
```sgquery
select:file lang:SCSS file:module patterntype:regexp archived:no fork:no
```


### Vulnerable and fixed Log4j versions
Confirm that vulnerable versions of log4j are removed and only fixed versions appear
```sgquery
lang:gradle org\.apache\.logging\.log4j['"] 2\.(0|1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16)(\.[0-9]+) patterntype:regexp archived:no fork:no
```
```sgquery
lang:gradle org\.apache\.logging\.log4j['"] 2\.(17)(\.[0-9]+) patterntype:regexp archived:no fork:no
```


### Yarn adoption
Are more repos increasingly using yarn? Track yarn adoption across teams and groups in your organization
```sgquery
select:repo file:yarn.lock archived:no fork:no
```


### Java versions
Detect and track which Java versions are most popular in your codebase


*Uses the [detect and track](../explanations/automatically_generated_data_series.md) capture groups insight type*
```sgquery
file:pom\.xml$ <java\.version>(.*)</java\.version> archived:no fork:no
```


### Linter override rules
A code health indicator for how many linter override rules exist
```sgquery
file:^\.eslintignore .\n patternType:regexp archived:no fork:no
```


### Language use over time
Track the growth of certain languages by file count
```sgquery
select:file lang:TypeScript
```
```sgquery
select:file lang:JavaScript
```

## Migration


### Config or docs file
How many repos contain a config or docs file in a specific directory
```sgquery
select:repo file:docs/*/new_config_filename archived:no fork:no
```


### “blacklist/whitelist” to “denylist/allowlist”
How the switch from files containing “blacklist/whitelist” to “denylist/allowlist” is progressing
```sgquery
select:file blacklist OR whitelist archived:no fork:no
```
```sgquery
select:file denylist OR allowlist archived:no fork:no
```


### Global CSS to CSS modules
Tracking migration from global CSS to CSS modules
```sgquery
select:file lang:SCSS -file:module patterntype:regexp archived:no fork:no
```
```sgquery
select:file lang:SCSS file:module patterntype:regexp archived:no fork:no
```


### Python 2 to Python 3
How far along is the Python major version migration
```sgquery
#!/usr/bin/env python3 archived:no fork:no
```
```sgquery
#!/usr/bin/env python2 archived:no fork:no
```


### React Class to Function Components Migration
What's the status of migrating to React function components from class components
```sgquery
patternType:regexp const\s\w+:\s(React\.)?FunctionComponent
```
```sgquery
patternType:regexp extends\s(React\.)?(Pure)?Component
```

## Adoption


### New API usage
How many repos or teams are using a new API your team built
```sgquery
select:repo ourApiLibraryName.load archived:no fork:no
```


### Yarn adoption
Are more repos increasingly using yarn? Track yarn adoption across teams and groups in your organization
```sgquery
select:repo file:yarn.lock archived:no fork:no
```


### Frequently used databases
Which databases we are calling or writing to most often
```sgquery
redis\.set patternType:regexp archived:no fork:no
```
```sgquery
graphql\( patternType:regexp archived:no fork:no
```


### Large or expensive package usage
Understand if a growing number of repos import a large/expensive package
```sgquery
select:repo import\slargePkg patternType:regexp archived:no fork:no
```


### React Component use
How many places are importing components from a library
```sgquery
from '@sourceLibrary/component' patternType:literal archived:no fork:no
```


### CI tooling adoption
How many repos are using our CI system
```sgquery
file:\.circleci/config.yml select:repo fork:no archived:no
```

## Deprecation


### CSS class
The removal of all deprecated CSS class
```sgquery
deprecated-class archived:no fork:no
```


### Icon or image
The removal of all deprecated icon or image instances
```sgquery
2018logo.png archived:no fork:no
```


### Structural code pattern
Deprecating a structural code pattern in favor of a safer pattern, like how many tries don't have catches
```sgquery
try {:[_]} catch (:[e]) { } finally {:[_]} lang:java patternType:structural archived:no fork:no
```


### Tooling
The progress of deprecating tooling you’re moving off of
```sgquery
deprecatedEventLogger.log archived:no fork:no
```


### Var keywords
Number of var keywords in the code basee (ES5 depreciation)
```sgquery
(lang:TypeScript OR lang:JavaScript) var ... = archived:no fork:no patterntype:structural
```


### Consolidation of Testing Libraries
Which React test libraries are being consolidated
```sgquery
from '@testing-library/react' archived:no fork:no
```
```sgquery
from 'enzyme' archived:no fork:no
```

## Versions and patterns
These examples are all for use with the [automatically generated data series](../explanations/automatically_generated_data_series.md) of "Detect and track" Code Insights, using regular expression capture groups.




### Java versions
Detect and track which Java versions are most popular in your codebase
```sgquery
file:pom\.xml$ <java\.version>(.*)</java\.version> archived:no fork:no
```


### License types in the codebase
See the breakdown of licenses from package.json files
```sgquery
file:package.json "license":\s"(.*)" archived:no fork:no
```


### All log4j versions
Which log4j versions are present, including vulnerable versions
```sgquery
lang:gradle org\.apache\.logging\.log4j['"] 2\.([0-9]+)\. archived:no fork:no
```


### Python versions
Which python versions are in use or haven’t been updated
```sgquery
#!/usr/bin/env python([0-9]\.[0-9]+) archived:no fork:no
```


### Node.js versions
Which node.js versions are present based on nvm files
```sgquery
nvm\suse\s([0-9]+\.[0-9]+) archived:no fork:no
```


### CSS Colors
What CSS colors are present or most popular
```sgquery
color:#([0-9a-fA-f]{3,6}) archived:no fork:no
```


### Types of checkov skips
See the most common reasons for why secuirty checks in checkov are skipped
```sgquery
patterntype:regexp file:.tf #checkov:skip=(.*) archived:no fork:no
```

## Code health


### TODOs
How many TODOs are in a specific part of the codebase (or all of it)
```sgquery
TODO archived:no fork:no
```


### Linter override rules
A code health indicator for how many linter override rules exist
```sgquery
file:^\.eslintignore .\n patternType:regexp archived:no fork:no
```


### Commits with “revert”
How frequently there are commits with “revert” in the commit message
```sgquery
type:commit revert archived:no fork:no
```


### Deprecated calls
How many times deprecated calls are used
```sgquery
lang:java @deprecated archived:no fork:no
```


### Storybook tests
How many tests for Storybook exist
```sgquery
patternType:regexp f:\.story\.tsx$ \badd\( archived:no fork:no
```


### Repos with Documentation
How many repos do or don't have READMEs
```sgquery
repohasfile:readme.md select:repo archived:no fork:no
```
```sgquery
-repohasfile:readme.md select:repo archived:no fork:no
```


### Ownership via CODEOWNERS files
How many repos do or don't have CODEOWNERS files
```sgquery
repohasfile:CODEOWNERS select:repo archived:no fork:no
```
```sgquery
-repohasfile:CODEOWNERS select:repo archived:no fork:no
```


### CI tooling adoption
How many repos are using our CI system
```sgquery
file:\.circleci/config.yml select:repo fork:no archived:no
```

## Security


### Vulnerable open source library
Confirm that a vulnerable open source library has been fully removed, or see the speed of the deprecation
```sgquery
vulnerableLibrary@14.3.9 archived:no fork:no
```


### API keys
How quickly we notice and remove API keys when they are committed
```sgquery
regexMatchingAPIKey patternType:regexp archived:no fork:no
```


### Vulnerable and fixed Log4j versions
Confirm that vulnerable versions of log4j are removed and only fixed versions appear
```sgquery
lang:gradle org\.apache\.logging\.log4j['"] 2\.(0|1|2|3|4|5|6|7|8|9|10|11|12|13|14|15|16)(\.[0-9]+) patterntype:regexp archived:no fork:no
```
```sgquery
lang:gradle org\.apache\.logging\.log4j['"] 2\.(17)(\.[0-9]+) patterntype:regexp archived:no fork:no
```


### How many tests are skipped
See how many tests have skip conditions
```sgquery
(this.skip() OR it.skip) lang:TypeScript archived:no fork:no
```


### Tests amount and types
See what types of tests are most common and total counts
```sgquery
patternType:regexp case:yes \b(it|test)\( f:/end-to-end/.*\.test\.ts$ archived:no fork:no
```
```sgquery
patternType:regexp case:yes \b(it|test)\( f:/regression/.*\.test\.ts$ archived:no fork:no
```
```sgquery
patternType:regexp case:yes \b(it|test)\( f:/integration/.*\.test\.ts$ archived:no fork:no
```


### Types of checkov skips
See the most common reasons for why secuirty checks in checkov are skipped


*Uses the [detect and track](../explanations/automatically_generated_data_series.md) capture groups insight type*
```sgquery
patterntype:regexp file:.tf #checkov:skip=(.*) archived:no fork:no
```

## Other


### Typescript vs. Go
Are there more Typescript or more Go files
```sgquery
select:file lang:TypeScript archived:no fork:no
```
```sgquery
select:file lang:Go archived:no fork:no
```


### iOS app screens
What number of iOS app screens are in the entire app
```sgquery
struct\s(.*):\sview$ patternType:regexp lang:swift archived:no fork:no
```


### Adopting new API by Team
Which teams or repos have adopted a new API so far
```sgquery
file:mobileTeam newAPI.call archived:no fork:no
```
```sgquery
file:webappTeam newAPI.call archived:no fork:no
```
*Or [filter teams by repositories](../how-tos/filtering_an_insight.md) in real time*


### Problematic API by Team
Which teams have the most usage of a problematic API
```sgquery
problemAPI file:teamOneDirectory archived:no fork:no
```
```sgquery
problemAPI file:teamTwoDirectory archived:no fork:no
```
*Or [filter teams by repositories](../how-tos/filtering_an_insight.md) in real time*


### Data fetching from GraphQL
What GraphQL operations are being called often
```sgquery
patternType:regexp requestGraphQL(\(|<[^>]*>\() archived:no fork:no
```
```sgquery
patternType:regexp (query|mutate)GraphQL(\(|<[^>]*>\() archived:no fork:no
```
```sgquery
patternType:regexp use(Query|Mutation|Connection|LazyQuery)(\(|<[^>]*>\() archived:no fork:no
```