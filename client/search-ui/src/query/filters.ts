import { FilterType } from '@sourcegraph/shared/src/search/query/filters'

export const filterDescriptions: Record<FilterType, string> = {
    [FilterType.after]:
        'Only include results from diffs or commits which have a commit date after the specified time frame. To use this filter, the search query must contain `type:diff` or `type:commit`.',
    [FilterType.archived]:
        'The "yes" option includes archived repositories. The "only" option filters results to only archived repositories. Results in archived repositories are excluded by default.',
    [FilterType.author]: `Only include results from diffs or commits authored by the user. Regexps are supported. Note that they match the whole author string of the form \`Full Name <user@example.com>\`, so to include only authors from a specific domain, use \`author:example.com>$\`.

You can also search by \`committer:git-email\`. *Note: there is a committer only when they are a different user than the author.*

To use this filter, the search query must contain \`type:diff\` or \`type:commit\`.`,
    [FilterType.before]:
        'Only include results from diffs or commits which have a commit date before the specified time frame. To use this filter, the search query must contain `type:diff` or `type:commit`.',
    [FilterType.case]: 'Perform a case sensitive query. Without this, everything is matched case insensitively.',
    [FilterType.committer]: `**Note:** There is a committer only when they are a different user than the author.

Only include results from diffs or commits commited by the user. Regexps are supported. They match the whole commiter string of the form \`Full Name <user@example.com>\`, so to include only commiters from a specific domain, use \`committer:example.com>$\`.

To use this filter, the search query must contain \`type:diff\` or \`type:commit\`.`,
    [FilterType.content]:
        'Set the search pattern with a dedicated parameter. Useful when searching literally for a string that may conflict with the search pattern syntax. In between the quotes, the `\\` character will need to be escaped (`\\\\` to evaluate for `\\`).',
    [FilterType.count]:
        'Retrieve *N* results. By default, Sourcegraph stops searching early and returns if it finds a full page of results. This is desirable for most interactive searches. To wait for all results, use **count:all**.',
    [FilterType.file]: 'Only include results in files whose full path matches the regexp.',
    [FilterType.fork]:
        'Include results from repository forks or filter results to only repository forks. Results in repository forks are exluded by default.',
    [FilterType.lang]: 'Only include results from files in the specified programming language.',
    [FilterType.message]: `Only include results from diffs or commits which have commit messages containing the string.

To use this filter, the search query must contain \`type:diff\` or \`type:commit\`.`,
    [FilterType.repo]:
        'Only include results from repositories whose path matches the regexp-pattern. A repository’s path is a string such as *github.com/myteam/abc* or *code.example.com/xyz* that depends on your organization’s repository host. If the regexp ends in `@rev`, that revision is searched instead of the default branch (usually `master`). `repo:regexp-pattern@rev` is equivalent to `repo:regexp-pattern rev:rev`.',
    [FilterType.rev]:
        'Search a revision instead of the default branch. `rev:` can only be used in conjunction with `repo:` and may not be used more than once. See our [revision syntax documentation](https://docs.sourcegraph.com/code_search/reference/queries#repository-revisions) to learn more.',
    [FilterType.select]: `Shows only query results for a given type. For example, \`select:repo\` displays only distinct repository paths from search results. The following values are available:

- \`select:repo\`
- \`select:commit.diff.added\`
- \`select:commit.diff.removed\`
- \`select:file\`
- \`select:file.directory\`
- \`select:file.path\`
- \`select:content\`
- \`select:symbol.symboltype\`

See [language definition](https://docs.sourcegraph.com/code_search/reference/language#select) for more information on possible values.`,
    [FilterType.type]:
        'Specifies the type of search. By default, searches are executed on all code at a given point in time (a branch or a commit). Specify the `type:` if you want to search over changes to code or commit messages instead (diffs or commits).',
    [FilterType.timeout]:
        'Customizes the timeout for searches. The value of the parameter is a string that can be parsed by the [Go time package’s `ParseDuration`](https://golang.org/pkg/time/#ParseDuration) (e.g. 10s, 100ms). By default, the timeout is set to 10 seconds, and the search will optimize for returning results as soon as possible. The timeout value cannot be set longer than 1 minute. When provided, the search is given the full timeout to complete.',
    [FilterType.visibility]:
        'Filter results to only public or private repositories. The default is to include both private and public repositories.',

    // These are deprecated or usually not provided manually
    [FilterType.context]: '',
    [FilterType.patterntype]: '',
    [FilterType.repohasfile]: '',
    [FilterType.repogroup]: '',
    [FilterType.repohascommitafter]: '',
}
