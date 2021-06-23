import React, { ReactElement, useCallback, useMemo, useState } from 'react'
import classNames from 'classnames'
import { Collapse } from 'reactstrap'
import { Tab, TabList, TabPanel, TabPanelProps, TabPanels, Tabs, useTabsContext } from '@reach/tabs'

import { VersionContextProps } from '@sourcegraph/shared/src/search/util'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../../..'
import { QueryChangeSource, QueryState } from '../../../helpers'

import styles from './SearchReference.module.scss'
import sidebarStyles from './SearchSidebarSection.module.scss'
import { FILTERS, FilterType, filterTypeKeys, NegatableFilter } from '@sourcegraph/shared/src/search/query/filters'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { Selection } from 'monaco-editor'
import { debounceTime, distinctUntilChanged, map, tap } from 'rxjs/operators'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { useEventObservable } from '@sourcegraph/shared/src/util/useObservable'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql/schema'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'
import { escapeRegExp } from 'lodash'

const SEARCH_REFERENCE_TAB_KEY = 'SearchProduct.SearchReference.Tab'

interface SearchReferenceInfo {
    type: FilterType
    placeholder: string
    prefix?: string
    suffix?: string
    value?: string
    description?: string
    showSuggestions?: boolean
    common?: true
}

const searchReferenceInfo: SearchReferenceInfo[] = [
    {
        type: FilterType.after,
        placeholder: '"last week"',
        description: `Only include results from diffs or commits which have a commit date after the specified time frame`,
        common: true,
    },
    {
        type: FilterType.archived,
        placeholder: 'yes or only',
        description: `The "yes" option includes archived repositories. The "only" option filters results to only archived repositories. Results in archived repositories are excluded by default.`,
    },
    {
        type: FilterType.case,
        placeholder: 'yes',
    },
    {
        type: FilterType.content,
        placeholder: '"pattern"',
        common: true,
    },
    {
        type: FilterType.count,
        placeholder: 'N or all',
        common: true,
    },
    {
        type: FilterType.file,
        placeholder: 'regexp-pattern',
        common: true,
    },
    {
        type: FilterType.fork,
        placeholder: 'yes or only',
        common: true,
    },
    {
        type: FilterType.lang,
        placeholder: 'language-name',
        common: true,
    },
    {
        type: FilterType.repo,
        placeholder: 'regexp-pattern',
        common: true,
    },
    {
        type: FilterType.repogroup,
        placeholder: 'group-name',
    },
    {
        type: FilterType.repo,
        placeholder: 'time',
        prefix: 'contains.commit.after(',
        suffix: ')',
        showSuggestions: false,
    },
    {
        type: FilterType.repo,
        placeholder: 'some content',
        prefix: 'contains(',
        suffix: ')',
        showSuggestions: false,
    },
    {
        type: FilterType.rev,
        placeholder: 'revision',
        common: true,
    },
    {
        type: FilterType.select,
        placeholder: 'result-types',
        common: true,
    },
    {
        type: FilterType.stable,
        placeholder: 'yes',
    },
    {
        type: FilterType.type,
        placeholder: 'symbol',
        common: true,
    },
    {
        type: FilterType.timeout,
        placeholder: 'golang-duration-value',
    },
    {
        type: FilterType.visibility,
        placeholder: 'any',
    },
]
const commonFilters = searchReferenceInfo.filter(info => info.common)

/**
 * Returns true if the provided regular expressions all match the provided
 * filter information (name, description, ...)
 */
function matches(searchTerms: RegExp[], info: SearchReferenceInfo): boolean {
    return searchTerms.every(term => {
        return term.test(info.type) || term.test(info.description || '')
    })
}

function parseSearchInput(searchInput: string): RegExp[] {
    const terms = searchInput.split(/\s+/)
    return terms.map(term => new RegExp(`\\b${escapeRegExp(term)}\\b`))
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
                        <span className="search-filter-keyword">
                            {searchReference.type}:{searchReference.prefix ?? ''}
                        </span>
                        {searchReference.value ?? (
                            <span className={styles.placeholder}>{searchReference.placeholder}</span>
                        )}
                    </span>
                    {searchReference.suffix ? (
                        <span className="search-filter-keyword">{searchReference.suffix}</span>
                    ) : null}
                </button>
            </span>
            <Collapse isOpen={!collapsed}>
                <div className={styles.description}>{searchReference.description}</div>
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
            {filters.map((filterInfo, i) => {
                return <SearchReferenceEntry searchReference={filterInfo} key={i} onClick={onClick} />
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
        (info: SearchReferenceInfo) => {
            // TODO: Do not just blindly append the filter. Possibly parse the
            // query and reuse existing filters if possible/necessary.

            const newQueryState: QueryState = {
                query: `${props.navbarSearchQueryState.query} ${info.type}:`,
                changeSource: QueryChangeSource.searchReference,
            }

            if (info.prefix) {
                newQueryState.query += info.prefix
            }

            if (info.value != null) {
                newQueryState.query += info.value
            } else if (FILTERS[info.type].discreteValues && info.showSuggestions !== false) {
                newQueryState.showSuggestions = true
            } else {
                let placeholder = info.placeholder
                let selectionStartPosition = newQueryState.query.length + 1
                let selectionEndPosition = selectionStartPosition + placeholder.length

                if (placeholder[0] === '"') {
                    selectionStartPosition += 1
                    selectionEndPosition -= 1
                }
                newQueryState.query += placeholder
                newQueryState.selection = new Selection(1, selectionStartPosition, 1, selectionEndPosition)
            }

            if (info.suffix) {
                newQueryState.query += info.suffix
            }

            props.onNavbarQueryChange(newQueryState)
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
                        <Tab>All filters</Tab>
                        <Tab>Common</Tab>
                        <Tab>Operators</Tab>
                    </TabList>
                    <TabPanels>
                        <TabPanel>{filterList}</TabPanel>
                        <TabPanel>
                            <SearchReferenceList filters={commonFilters} onClick={updateQuery} />
                        </TabPanel>
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
): (filter: string) => React.ReactElement {
    return (filter: string) => <SearchReference {...props} filter={filter} />
}
