import React, { ReactElement, useCallback, useMemo, useState } from 'react'
import classNames from 'classnames'
import { Collapse } from 'reactstrap'
import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'

import { VersionContextProps } from '@sourcegraph/shared/src/search/util'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../../..'
import { QueryChangeSource, QueryState } from '../../../helpers'

import styles from './SearchReference.module.scss'
import sidebarStyles from './SearchSidebarSection.module.scss'
import { FILTERS, FilterType } from '@sourcegraph/shared/src/search/query/filters'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { Selection } from 'monaco-editor'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'
import { escapeRegExp } from 'lodash'
import { updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { findFilter, FilterKind } from '@sourcegraph/shared/src/search/query/validate'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { renderMarkdown } from '@sourcegraph/shared/src/util/markdown'

const SEARCH_REFERENCE_TAB_KEY = 'SearchProduct.SearchReference.Tab'

interface SearchReferenceInfo {
    type: FilterType
    placeholder: Placeholder
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
}

const searchReferenceInfo: SearchReferenceInfo[] = [
    {
        type: FilterType.after,
        placeholder: parsePlaceholder('"{last week}"'),
        description: `Only include results from diffs or commits which have a commit date after the specified time frame.`,
        commonRank: 100,
    },
    {
        type: FilterType.archived,
        placeholder: parsePlaceholder('{yes or only}'),
        description: `The "yes" option includes archived repositories. The "only" option filters results to only archived repositories. Results in archived repositories are excluded by default.`,
    },
    {
        type: FilterType.case,
        placeholder: parsePlaceholder('{yes}'),
        description: `Perform a case sensitive query. Without this, everything is matched case insensitively.`,
    },
    {
        type: FilterType.content,
        placeholder: parsePlaceholder('"{pattern}"'),
        description:
            'Set the search pattern with a dedicated parameter. Useful when searching literally for a string that may conflict with the [search pattern syntax](https://docs.sourcegraph.com/code_search/reference/queries#search-pattern-syntax). In between the quotes, the `\\` character will need to be escaped (`\\\\` to evaluate for `\\`).',
        commonRank: 70,
    },
    {
        type: FilterType.count,
        placeholder: parsePlaceholder('{N or all}'),
        description:
            'Retrieve *N* results. By default, Sourcegraph stops searching early and returns if it finds a full page of results. This is desirable for most interactive searches. To wait for all results, use **count:all**.',
        commonRank: 60,
    },
    {
        type: FilterType.file,
        placeholder: parsePlaceholder('{regexp-pattern}'),
        commonRank: 30,
        description: 'Only include results in files whose full path matches the regexp.',
    },
    {
        type: FilterType.fork,
        placeholder: parsePlaceholder('{yes or only}'),
        description:
            'Include results from repository forks or filter results to only repository forks. Results in repository forks are exluded by default.',
        commonRank: 80,
    },
    {
        type: FilterType.lang,
        placeholder: parsePlaceholder('{language-name}'),
        description: 'Only include results from files in the specified programming language.',
        commonRank: 40,
    },
    {
        type: FilterType.repo,
        placeholder: parsePlaceholder('{regexp-pattern}'),
        description:
            'Only include results from repositories whose path matches the regexp-pattern. A repository’s path is a string such as *github.com/myteam/abc* or *code.example.com/xyz* that depends on your organization’s repository host. If the regexp ends in [`@rev`](https://docs.sourcegraph.com/code_search/reference/queries#repository-revisions), that revision is searched instead of the default branch (usually `master`). `repo:regexp-pattern@rev` is equivalent to `repo:regexp-pattern rev:rev`.',
        commonRank: 10,
    },
    {
        type: FilterType.repogroup,
        placeholder: parsePlaceholder('{group-name}'),
    },
    {
        type: FilterType.repo,
        placeholder: parsePlaceholder('contains.commit.after({time})'),
        showSuggestions: false,
    },
    {
        type: FilterType.repo,
        placeholder: parsePlaceholder('contains({file:foo content:bar})'),
        showSuggestions: false,
    },
    {
        type: FilterType.rev,
        placeholder: parsePlaceholder('{revision}'),
        commonRank: 20,
    },
    {
        type: FilterType.select,
        placeholder: parsePlaceholder('{result-types}'),
        commonRank: 50,
    },
    {
        type: FilterType.stable,
        placeholder: parsePlaceholder('{yes}'),
    },
    {
        type: FilterType.type,
        placeholder: parsePlaceholder('{symbol}'),
        commonRank: 90,
    },
    {
        type: FilterType.timeout,
        placeholder: parsePlaceholder('{golang-duration-value}'),
    },
    {
        type: FilterType.visibility,
        placeholder: parsePlaceholder('{any}'),
    },
]

const commonFilters = searchReferenceInfo
    .filter(info => info.commonRank != null)
    .sort((a, b) => a.commonRank - b.commonRank)

/**
 * Returns true if the provided regular expressions all match the provided
 * filter information (name, description, ...)
 */
function matches(searchTerms: RegExp[], info: SearchReferenceInfo): boolean {
    return searchTerms.every(term => {
        return term.test(info.type) || term.test(info.description || '')
    })
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
 * Given a Placeholder object, this function returns the Monaco Selections
 * corresponding to each value in the Placeholder.
 */
function selectionsForPlaceholder(placeholder: Placeholder, offset: number = 0): Selection[] {
    return placeholder.tokens
        .filter(token => token.type === 'value')
        .map(token => new Selection(1, offset + token.start, 1, offset + token.end - 1))
}

/**
 * Whether or not to trigger the suggestion popover when adding this filter to
 * the query.
 */
function showSuggestions(searchReference: SearchReferenceInfo): boolean {
    return Boolean(searchReference.showSuggestions !== false && FILTERS[searchReference.type].discreteValues)
}

/**
 * This helper function will update the current query with the provided filter,
 * updating existing filters if necessary.
 */
function updateQueryWithFilter(currentQueryState: QueryState, searchReference: SearchReferenceInfo): QueryState {
    // If a filter has suggestions, we will always add or replace it with an
    // empty value and trigger the suggestion popover
    if (showSuggestions(searchReference)) {
        return {
            query: updateFilter(currentQueryState.query, searchReference.type, ''),
            showSuggestions: true,
            changeSource: QueryChangeSource.searchReference,
        }
    }

    const existingFilter = findFilter(currentQueryState.query, searchReference.type, FilterKind.Global)

    // If the filter doesn't exist yet we simply append it with a placeholder
    // value and select that value
    if (!existingFilter) {
        const query = updateFilter(currentQueryState.query, searchReference.type, searchReference.placeholder.text)
        return {
            query,
            changeSource: QueryChangeSource.searchReference,
            selection: selectionsForPlaceholder(
                searchReference.placeholder,
                query.length - searchReference.placeholder.text.length + 1
            )[0],
        }
    }

    // Otherwise we just select the existing value
    return {
        query: currentQueryState.query,
        changeSource: QueryChangeSource.searchReference,
        selection: new Selection(
            1,
            (existingFilter.value?.range.start || existingFilter.field.range.end) + 1,
            1,
            existingFilter.range.end + 1
        ),
    }
}

interface Placeholder {
    tokens: Array<{ type: 'text' | 'value'; content: string; start: number; end: number }>
    text: string
}

function parsePlaceholder(placeholder: string): Placeholder {
    const valuePattern = /\{([^}]+)\}/g
    let currentIndex = 0
    let parsedPlaceholder: Placeholder = { tokens: [], text: '' }
    let match
    while ((match = valuePattern.exec(placeholder))) {
        if (currentIndex !== match.index) {
            parsedPlaceholder.tokens.push({
                type: 'text',
                content: placeholder.slice(currentIndex, match.index),
                start: currentIndex,
                end: match.index,
            })
        }
        parsedPlaceholder.tokens.push({
            type: 'value',
            content: match[1],
            start: match.index,
            end: match[0].length,
        })
        currentIndex = match.index + match[0].length
    }

    if (currentIndex < placeholder.length) {
        parsedPlaceholder.tokens.push({
            type: 'text',
            content: placeholder.slice(currentIndex),
            start: currentIndex,
            end: placeholder.length,
        })
    }
    parsedPlaceholder.text = parsedPlaceholder.tokens.map(token => token.content).join('')
    return parsedPlaceholder
}

const classNameTokenMap = {
    text: 'search-filter-keyword',
    value: styles.placeholder,
}

interface SearchReferenceEntryProps {
    searchReference: SearchReferenceInfo
    onClick: (searchReference: SearchReferenceInfo) => void
}

const SearchReferenceEntry: React.FunctionComponent<SearchReferenceEntryProps> = ({ searchReference, onClick }) => {
    const [collapsed, setCollapsed] = useState(true)
    const CollapseIcon = collapsed ? ChevronRightIcon : ChevronDownIcon

    return (
        <li>
            <span
                className={classNames(styles.item, sidebarStyles.sidebarSectionListItem, {
                    [styles.active]: !collapsed,
                })}
            >
                <button
                    type="button"
                    className={classNames('btn btn-icon mr-1', styles.collapseButton)}
                    onClick={event => {
                        event.stopPropagation()
                        setCollapsed(collapsed => !collapsed)
                    }}
                    aria-label={collapsed ? 'Show filter description' : 'Hide filter description'}
                >
                    <CollapseIcon className="icon-inline" />
                </button>
                <button className="btn p-0" type="button" onClick={() => onClick(searchReference)}>
                    <span className="text-monospace">
                        <span className="search-filter-keyword">{searchReference.type}:</span>
                        {searchReference.placeholder.tokens.map(token => (
                            <span key={token.start} className={classNameTokenMap[token.type]}>
                                {token.content}
                            </span>
                        ))}
                    </span>
                </button>
            </span>
            <Collapse isOpen={!collapsed}>
                <div className={styles.description}>
                    {searchReference.description && (
                        <Markdown dangerousInnerHTML={renderMarkdown(searchReference.description)} />
                    )}
                </div>
            </Collapse>
        </li>
    )
}

interface SearchReferenceListProps {
    filters: SearchReferenceInfo[]
    onClick: (info: SearchReferenceInfo) => void
}

const SearchReferenceList = ({ filters, onClick }: SearchReferenceListProps): ReactElement => {
    return (
        <ul className={styles.list}>
            {filters.map(filterInfo => {
                return (
                    <SearchReferenceEntry
                        searchReference={filterInfo}
                        key={filterInfo.type + filterInfo.placeholder.text}
                        onClick={onClick}
                    />
                )
            })}
        </ul>
    )
}

export interface SearchReferenceProps
    extends PatternTypeProps,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    query: string
    filter: string
    navbarSearchQueryState: QueryState
    onNavbarQueryChange: (queryState: QueryState) => void
}

const SearchReference = (props: SearchReferenceProps): ReactElement => {
    const [selectedTab, setSelectedTab] = useLocalStorage(SEARCH_REFERENCE_TAB_KEY, 0)

    const selectedFilters = useMemo(() => {
        if (props.filter === '') {
            return searchReferenceInfo
        }
        const searchTerms = parseSearchInput(props.filter)
        return searchReferenceInfo.filter(info => matches(searchTerms, info))
    }, [props.filter])

    const updateQuery = useCallback(
        (searchReference: SearchReferenceInfo) => {
            props.onNavbarQueryChange(updateQueryWithFilter(props.navbarSearchQueryState, searchReference))
        },
        [props.onNavbarQueryChange, props.navbarSearchQueryState]
    )

    const filterList = <SearchReferenceList filters={selectedFilters} onClick={updateQuery} />

    return (
        <div>
            {props.filter ? (
                filterList
            ) : (
                <Tabs index={selectedTab} onChange={setSelectedTab}>
                    <TabList className={styles.tablist}>
                        <Tab>Common</Tab>
                        <Tab>All filters</Tab>
                        <Tab>Operators</Tab>
                    </TabList>
                    <TabPanels>
                        <TabPanel>
                            <SearchReferenceList filters={commonFilters} onClick={updateQuery} />
                        </TabPanel>
                        <TabPanel>{filterList}</TabPanel>
                        <TabPanel>TODO</TabPanel>
                    </TabPanels>
                </Tabs>
            )}
            <p className={styles.footer}>
                <Link to="https://docs.sourcegraph.com/code_search/reference/queries">Search syntax</Link>
            </p>
        </div>
    )
}

export function getSearchReferenceFactory(
    props: Omit<SearchReferenceProps, 'filter'>
): (filter: string) => ReactElement {
    return (filter: string) => <SearchReference {...props} filter={filter} />
}
