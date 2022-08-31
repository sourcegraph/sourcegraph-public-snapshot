import React, { ReactElement, useCallback, useMemo, useState } from 'react'

import { mdiChevronDown, mdiChevronLeft, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'
import { escapeRegExp } from 'lodash'

import { renderMarkdown } from '@sourcegraph/common'
import {
    SearchQueryState,
    createQueryExampleFromString,
    updateQueryWithFilterAndExample,
    QueryExample,
    EditorHint,
} from '@sourcegraph/search'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { FILTERS, FilterType, isNegatableFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Button,
    useLocalStorage,
    Link,
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    Tabs,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    Icon,
    Text,
} from '@sourcegraph/wildcard'

import sidebarStyles from './SearchFilterSection.module.scss'
import styles from './SearchReference.module.scss'

const SEARCH_REFERENCE_TAB_KEY = 'SearchProduct.SearchReference.Tab'

type FilterInfo = QueryExample & {
    field: FilterType
    description: string
    /**
     * Force showing or not showing suggestions for this fileter.
     */
    showSuggestions?: boolean
    /**
     * Used to indicate whether this filter/example should be listed in the
     * "Common" filters section and at which position
     */
    commonRank?: number
    alias?: string
    examples?: string[]
}

type OperatorInfo = QueryExample & {
    operator: string
    description: string
    alias?: string
    examples?: string[]
}

type SearchReferenceInfo = FilterInfo | OperatorInfo

/**
 * Adds additional search reference information from the existing filters list.
 */
function augmentFilterInfo(searchReference: FilterInfo): void {
    const filter = FILTERS[searchReference.field]
    if (filter?.alias) {
        searchReference.alias = filter.alias
    }
}

const filterInfos: FilterInfo[] = [
    {
        ...createQueryExampleFromString('"{last week}"'),
        field: FilterType.after,
        description:
            'Only include results from diffs or commits which have a commit date after the specified time frame. To use this filter, the search query must contain `type:diff` or `type:commit`.',
        commonRank: 100,
        examples: ['after:"6 weeks ago"', 'after:"november 1 2019"'],
    },
    {
        ...createQueryExampleFromString('{yes/only}'),
        field: FilterType.archived,
        description:
            'The "yes" option includes archived repositories. The "only" option filters results to only archived repositories. Results in archived repositories are excluded by default.',
        examples: ['repo:sourcegraph/ archived:only'],
    },
    {
        ...createQueryExampleFromString('{name}'),
        field: FilterType.author,
        description: `Only include results from diffs or commits authored by the user. Regexps are supported. Note that they match the whole author string of the form \`Full Name <user@example.com>\`, so to include only authors from a specific domain, use \`author:example.com>$\`.

You can also search by \`committer:git-email\`. *Note: there is a committer only when they are a different user than the author.*

To use this filter, the search query must contain \`type:diff\` or \`type:commit\`.`,
        examples: ['type:diff author:nick'],
    },
    {
        ...createQueryExampleFromString('"{last thursday}"'),
        field: FilterType.before,
        description:
            'Only include results from diffs or commits which have a commit date before the specified time frame. To use this filter, the search query must contain `type:diff` or `type:commit`.',
        commonRank: 100,
        examples: ['before:"last thursday"', 'before:"november 1 2019"'],
    },
    {
        ...createQueryExampleFromString('{yes}'),
        field: FilterType.case,
        description: 'Perform a case sensitive query. Without this, everything is matched case insensitively.',
        examples: ['OPEN_FILE case:yes'],
    },
    {
        ...createQueryExampleFromString('"{pattern}"'),
        field: FilterType.content,
        description:
            'Set the search pattern with a dedicated parameter. Useful when searching literally for a string that may conflict with the search pattern syntax. In between the quotes, the `\\` character will need to be escaped (`\\\\` to evaluate for `\\`).',
        commonRank: 70,
        examples: ['repo:sourcegraph content:"repo:sourcegraph"', 'file:Dockerfile alpine -content:alpine:latest'],
    },
    {
        ...createQueryExampleFromString('{N/all}'),
        field: FilterType.count,
        description:
            'Retrieve *N* results. By default, Sourcegraph stops searching early and returns if it finds a full page of results. This is desirable for most interactive searches. To wait for all results, use **count:all**.',
        commonRank: 60,
        examples: ['count:1000 function', 'count:all err'],
    },
    {
        ...createQueryExampleFromString('{regexp-pattern}'),
        field: FilterType.file,
        commonRank: 30,
        description: 'Only include results in files whose full path matches the regexp.',
        examples: ['file:.js$ httptest', 'file:internal/ httptest', 'file:.js$ -file:test http'],
    },
    {
        ...createQueryExampleFromString('has.content({regexp-pattern})'),
        field: FilterType.file,
        description: 'Search only inside files that contain content matching the provided regexp pattern.',
        examples: ['file:has.content(github.com/sourcegraph/sourcegraph)'],
    },
    {
        ...createQueryExampleFromString('{yes/only}'),
        field: FilterType.fork,
        description:
            'Include results from repository forks or filter results to only repository forks. Results in repository forks are exluded by default.',
        commonRank: 80,
        examples: ['fork:yes repo:sourcegraph'],
    },
    {
        ...createQueryExampleFromString('{language-name}'),
        field: FilterType.lang,
        description: 'Only include results from files in the specified programming language.',
        commonRank: 40,
        examples: ['lang:typescript encoding', '-lang:typescript encoding'],
    },
    {
        ...createQueryExampleFromString('"{any string}"'),
        field: FilterType.message,
        description: `Only include results from diffs or commits which have commit messages containing the string.

To use this filter, the search query must contain \`type:diff\` or \`type:commit\`.`,
        examples: ['type:commit message:"testing"', 'type:diff message:"testing"'],
    },
    {
        ...createQueryExampleFromString('{regexp-pattern}'),
        field: FilterType.repo,
        description:
            'Only include results from repositories whose path matches the regexp-pattern. A repository’s path is a string such as *github.com/myteam/abc* or *code.example.com/xyz* that depends on your organization’s repository host. If the regexp ends in `@rev`, that revision is searched instead of the default branch (usually `master`). `repo:regexp-pattern@rev` is equivalent to `repo:regexp-pattern rev:rev`.',
        commonRank: 10,
        examples: [
            'repo:gorilla/mux testroute',
            'repo:^github.com/sourcegraph/sourcegraph$@v3.14.0 mux',
            'repo:alice/ -repo:old-repo',
            'repo:vscode@*refs/heads/:^refs/heads/master type:diff task',
        ],
    },
    {
        ...createQueryExampleFromString('has.path({path})'),
        field: FilterType.repo,
        description: 'Search only inside repositories that contain a file path matching the regular expression.',
        examples: ['repo:has.path(README)', 'repo:has.path(src/main/)'],
        showSuggestions: false,
    },
    {
        ...createQueryExampleFromString('has.content({content})'),
        field: FilterType.repo,
        description: 'Search only inside repositories that contain file content matching the regular expression.',
        examples: ['repo:has.content(TODO)'],
        showSuggestions: false,
    },
    {
        ...createQueryExampleFromString('has.file({path:path content:content})'),
        field: FilterType.repo,
        description:
            'Search only inside repositories that contain a file path matching the `path:` and/or `content:` filters.',
        examples: ['repo:has.file(path:README content:fix)'],
        showSuggestions: false,
    },
    {
        ...createQueryExampleFromString('has.commit.after({date})'),
        field: FilterType.repo,
        description:
            'Search only inside repositories that contain a a commit after some specified time. See [git date formats](https://github.com/git/git/blob/master/Documentation/date-formats.txt) for accepted formats. Use this to filter out stale repositories that don’t contain commits past the specified time frame. This parameter is experimental.',
        examples: ['repo:has.commit.after(1 month ago)', 'repo:has.commit.after(june 25 2017)'],
        showSuggestions: false,
    },
    {
        ...createQueryExampleFromString('has.description({regex-pattern})'),
        field: FilterType.repo,
        description: 'Search inside repositories that have a description matched by the provided regex pattern.',
        examples: ['repo:has.description(linux kernel)', 'repo:has.description(go.*library)'],
        showSuggestions: false,
    },
    {
        ...createQueryExampleFromString('{revision}'),
        field: FilterType.rev,
        commonRank: 20,
        description:
            'Search a revision instead of the default branch. `rev:` can only be used in conjunction with `repo:` and may not be used more than once. See our [revision syntax documentation](https://docs.sourcegraph.com/code_search/reference/queries#repository-revisions) to learn more.',
    },
    {
        ...createQueryExampleFromString('{result-types}'),
        field: FilterType.select,
        commonRank: 50,
        description: `Shows only query results for a given type. For example, \`select:repo\` displays only distinct repository paths from search results. The following values are available:

- \`select:repo\`
- \`select:commit.diff.added\`
- \`select:commit.diff.removed\`
- \`select:file\`
- \`select:file.directory\`
- \`select:file.path\`
- \`select:content\`
- \`select:symbol.symboltype\`

See [language definition](https://docs.sourcegraph.com/code_search/reference/language#select) for more information on possible values.`,
        examples: ['fmt.Errorf select:repo', 'select:commit.diff.added //TODO', 'select:file.directory'],
    },
    {
        ...createQueryExampleFromString('{diff/commit/...}'),
        field: FilterType.type,
        commonRank: 1,
        description:
            'Specifies the type of search. By default, searches are executed on all code at a given point in time (a branch or a commit). Specify the `type:` if you want to search over changes to code or commit messages instead (diffs or commits).',
        examples: ['type:symbol path', 'type:diff func', 'type:commit test'],
    },
    {
        ...createQueryExampleFromString('{golang-duration-value}'),
        field: FilterType.timeout,
        description:
            'Customizes the timeout for searches. The value of the parameter is a string that can be parsed by the [Go time package’s `ParseDuration`](https://golang.org/pkg/time/#ParseDuration) (e.g. 10s, 100ms). By default, the timeout is set to 10 seconds, and the search will optimize for returning results as soon as possible. The timeout value cannot be set longer than 1 minute. When provided, the search is given the full timeout to complete.',
        examples: ['repo:^github.com/sourcegraph timeout:15s func count:10000'],
    },
    {
        ...createQueryExampleFromString('{any/private/public}'),
        field: FilterType.visibility,
        description:
            'Filter results to only public or private repositories. The default is to include both private and public repositories.',
        examples: ['type:repo visibility:public'],
    },
]

for (const info of filterInfos) {
    augmentFilterInfo(info)
}

const commonFilters = filterInfos
    .filter(info => info.commonRank !== undefined)
    // commonRank will never be undefined here, but TS doesn't seem to know
    .sort((a, b) => (a.commonRank as number) - (b.commonRank as number))

const operatorInfo: OperatorInfo[] = [
    {
        ...createQueryExampleFromString('{expr} AND {expr}'),
        operator: 'AND',
        alias: 'and',
        description:
            'Returns results for files containing matches on the left and right side of the `and` (set intersection).',
        examples: ['conf.Get( and log15.Error(', 'conf.Get( AND log15.Error( AND after'],
    },
    {
        ...createQueryExampleFromString('({expr} OR {expr})'),
        operator: 'OR',
        alias: 'or',
        description:
            'Returns file content matching either on the left or right side, or both (set union). The number of results reports the number of matches of both strings.',
        examples: ['conf.Get( or log15.Error(', 'conf.Get( OR log15.Error( OR after'],
    },
    {
        ...createQueryExampleFromString('NOT {expr}'),
        operator: 'NOT',
        alias: 'not',
        description:
            '`NOT` can be prepended to negate filters like `file`, `lang`, `repo`. Prepending `NOT` to search patterns excludes documents that contain the pattern. For readability, you may use `NOT` in conjunction with `AND` if you like: `panic AND NOT ever`.',
        examples: ['lang:go not file:main.go panic', 'panic NOT ever'],
    },
]

/**
 * Returns true if the provided regular expressions all match the provided
 * filter information (name, description, ...)
 */
function matches(searchTerms: RegExp[], info: FilterInfo): boolean {
    return searchTerms.every(term => term.test(info.field) || term.test(info.description || ''))
}

/**
 * Convert the search input into an array of regular expressions. Each word in
 * the input becomes a regular expression starting with a word boundary check.
 */
function parseSearchInput(searchInput: string): RegExp[] {
    const terms = searchInput.split(/\s+/)
    return terms.map(term => new RegExp(`\\b${escapeRegExp(term)}`))
}

/**
 * Whether or not to trigger the suggestion popover when adding this filter to
 * the query.
 */
function shouldShowSuggestions(searchReference: FilterInfo): boolean {
    return Boolean(searchReference.showSuggestions !== false && FILTERS[searchReference.field].discreteValues)
}

function isFilterInfo(searchReference: SearchReferenceInfo): searchReference is FilterInfo {
    return (searchReference as FilterInfo).field !== undefined
}

const classNameTokenMap = {
    text: 'search-filter-keyword',
    placeholder: styles.placeholder,
}

interface SearchReferenceExampleProps {
    example: string
    onClick?: (example: string) => void
}

const SearchReferenceExample: React.FunctionComponent<React.PropsWithChildren<SearchReferenceExampleProps>> = ({
    example,
    onClick,
}) => {
    const scanResult = scanSearchQuery(example, false, SearchPatternType.standard)
    // We only use valid queries as examples, so this will always be true
    if (scanResult.type === 'success') {
        return (
            <Button className="p-0 flex-1" onClick={() => onClick?.(example)}>
                {scanResult.term.map((term, index) => {
                    switch (term.type) {
                        case 'filter':
                            return (
                                <React.Fragment key={index}>
                                    <span className="search-filter-keyword">{term.field.value}:</span>
                                    {term.value?.quoted ? `"${term.value.value}"` : term.value?.value}
                                </React.Fragment>
                            )
                        case 'keyword':
                            return (
                                <span key={index} className="search-filter-keyword">
                                    {term.value}
                                </span>
                            )
                        default:
                            return example.slice(term.range.start, term.range.end)
                    }
                })}
            </Button>
        )
    }
    return null
}

interface SearchReferenceEntryProps<T extends SearchReferenceInfo> {
    searchReference: T
    onClick: (searchReference: T, negate: boolean) => void
    onExampleClick?: (example: string) => void
}

const SearchReferenceEntry = <T extends SearchReferenceInfo>({
    searchReference,
    onClick,
    onExampleClick,
}: SearchReferenceEntryProps<T>): ReactElement | null => {
    const [collapsed, setCollapsed] = useState(true)
    const collapseIcon = collapsed ? mdiChevronLeft : mdiChevronDown

    const handleOpenChange = useCallback((collapsed: boolean) => setCollapsed(!collapsed), [])

    let buttonTextPrefix: ReactElement | null = null
    if (isFilterInfo(searchReference)) {
        buttonTextPrefix = <span className="search-filter-keyword">{searchReference.field}:</span>
    }

    return (
        <li>
            <Collapse isOpen={!collapsed} onOpenChange={handleOpenChange}>
                <span
                    className={classNames(styles.item, sidebarStyles.sidebarSectionListItem, {
                        [styles.active]: !collapsed,
                    })}
                >
                    <Button className="p-0 flex-1" onClick={event => onClick(searchReference, event.altKey)}>
                        <span className="text-monospace">
                            {buttonTextPrefix}
                            {searchReference.tokens.map(token => (
                                <span key={token.start} className={classNameTokenMap[token.type]}>
                                    {token.value}
                                </span>
                            ))}
                        </span>
                    </Button>
                    <CollapseHeader
                        as={Button}
                        variant="icon"
                        className={styles.collapseButton}
                        aria-label={collapsed ? 'Show filter description' : 'Hide filter description'}
                    >
                        <small className="text-monospace">i</small>
                        <Icon aria-hidden={true} svgPath={collapseIcon} />
                    </CollapseHeader>
                </span>
                <CollapsePanel>
                    {!collapsed && (
                        <div className={styles.description}>
                            {searchReference.description && (
                                <Markdown dangerousInnerHTML={renderMarkdown(searchReference.description)} />
                            )}
                            {searchReference.alias && (
                                <Text>
                                    Alias:{' '}
                                    <span className="text-code search-filter-keyword">
                                        {searchReference.alias}
                                        {isFilterInfo(searchReference) ? ':' : ''}
                                    </span>
                                </Text>
                            )}
                            {isFilterInfo(searchReference) && isNegatableFilter(searchReference.field) && (
                                <Text>
                                    Negation:{' '}
                                    <span className="test-code search-filter-keyword">-{searchReference.field}:</span>
                                    {searchReference.alias && (
                                        <>
                                            {' '}
                                            |{' '}
                                            <span className="test-code search-filter-keyword">
                                                -{searchReference.alias}:
                                            </span>
                                        </>
                                    )}
                                    <br />
                                    <span className={styles.placeholder}>(opt + click filter in reference list)</span>
                                </Text>
                            )}
                            {searchReference.examples && (
                                <>
                                    <div className="font-weight-medium">Examples</div>
                                    <div className={classNames('text-code', styles.examples)}>
                                        {searchReference.examples.map(example => (
                                            <Text key={example}>
                                                <SearchReferenceExample example={example} onClick={onExampleClick} />
                                            </Text>
                                        ))}
                                    </div>
                                </>
                            )}
                        </div>
                    )}
                </CollapsePanel>
            </Collapse>
        </li>
    )
}

interface FilterInfoListProps<T extends SearchReferenceInfo> {
    filters: T[]
    onClick: (info: T, negate: boolean) => void
    onExampleClick: (example: string) => void
}

const FilterInfoList = ({ filters, onClick, onExampleClick }: FilterInfoListProps<FilterInfo>): ReactElement => (
    <ul className={styles.list}>
        {filters.map(filterInfo => (
            <SearchReferenceEntry
                key={filterInfo.field + filterInfo.value}
                searchReference={filterInfo}
                onClick={onClick}
                onExampleClick={onExampleClick}
            />
        ))}
    </ul>
)

export interface SearchReferenceProps extends TelemetryProps, Pick<SearchQueryState, 'setQueryState'> {
    filter: string
}

const SearchReference = React.memo(
    (props: SearchReferenceProps): ReactElement => {
        const [persistedTabIndex, setPersistedTabIndex] = useLocalStorage(SEARCH_REFERENCE_TAB_KEY, 0)

        const { setQueryState, telemetryService } = props
        const filter = props.filter.trim()
        const hasFilter = filter.length > 0

        const selectedFilters = useMemo(() => {
            if (!hasFilter) {
                return filterInfos
            }
            const searchTerms = parseSearchInput(filter)
            return filterInfos.filter(info => matches(searchTerms, info))
        }, [filter, hasFilter])

        const updateQuery = useCallback(
            (searchReference: FilterInfo, negate: boolean) => {
                setQueryState(({ query }) => {
                    const updatedQuery = updateQueryWithFilterAndExample(
                        query,
                        searchReference.field,
                        searchReference,
                        {
                            singular: Boolean(FILTERS[searchReference.field].singular),
                            negate: negate && isNegatableFilter(searchReference.field),
                            emptyValue: shouldShowSuggestions(searchReference),
                        }
                    )
                    return {
                        query: updatedQuery.query,
                        selectionRange: updatedQuery.placeholderRange,
                        revealRange: updatedQuery.filterRange,
                        hint:
                            (shouldShowSuggestions(searchReference) ? EditorHint.ShowSuggestions : 0) |
                            EditorHint.Focus,
                    }
                })
            },
            [setQueryState]
        )
        const updateQueryWithOperator = useCallback(
            (info: OperatorInfo) => {
                setQueryState(({ query }) => ({ query: query + ` ${info.operator} ` }))
            },
            [setQueryState]
        )
        const updateQueryWithExample = useCallback(
            (example: string) => {
                telemetryService.log(hasFilter ? 'SearchReferenceSearchedAndClicked' : 'SearchReferenceFilterClicked')
                setQueryState(({ query }) => ({ query: query.trimEnd() + ' ' + example }))
            },
            [setQueryState, hasFilter, telemetryService]
        )

        const filterList = (
            <FilterInfoList filters={selectedFilters} onClick={updateQuery} onExampleClick={updateQueryWithExample} />
        )

        return (
            <div>
                {hasFilter ? (
                    filterList
                ) : (
                    <Tabs className={styles.tabs} defaultIndex={persistedTabIndex} onChange={setPersistedTabIndex}>
                        <TabList>
                            <Tab>Common</Tab>
                            <Tab>All filters</Tab>
                            <Tab>Operators</Tab>
                        </TabList>
                        <TabPanels>
                            <TabPanel>
                                <FilterInfoList
                                    filters={commonFilters}
                                    onClick={updateQuery}
                                    onExampleClick={updateQueryWithExample}
                                />
                            </TabPanel>
                            <TabPanel>{filterList}</TabPanel>
                            <TabPanel>
                                <ul className={styles.list}>
                                    {operatorInfo.map(operatorInfo => (
                                        <SearchReferenceEntry
                                            searchReference={operatorInfo}
                                            key={operatorInfo.operator + operatorInfo.value}
                                            onClick={updateQueryWithOperator}
                                            onExampleClick={updateQueryWithExample}
                                        />
                                    ))}
                                </ul>
                            </TabPanel>
                        </TabPanels>
                    </Tabs>
                )}
                <Text className={sidebarStyles.sidebarSectionFooter}>
                    <small>
                        <Link target="blank" to="https://docs.sourcegraph.com/code_search/reference/queries">
                            Search syntax <Icon role="img" aria-label="Open in a new tab" svgPath={mdiOpenInNew} />
                        </Link>
                    </small>
                </Text>
            </div>
        )
    }
)

export function getSearchReferenceFactory(
    props: Omit<SearchReferenceProps, 'filter'>
): (filter: string) => React.ReactNode {
    return (filter: string) => <SearchReference {...props} filter={filter} />
}
