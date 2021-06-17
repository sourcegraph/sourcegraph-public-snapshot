import React, { ReactElement, useCallback, useState } from 'react'
import classNames from 'classnames'
import { Collapse } from 'reactstrap'

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

interface SearchReferenceInfo {
    placeholder?: string
    value?: string
}

const searchReferenceInfo: Record<Exclude<FilterType, NegatableFilter>, SearchReferenceInfo> &
    Record<NegatableFilter, SearchReferenceInfo> = {
    [FilterType.after]: {
        placeholder: '"string time"',
    },
    [FilterType.before]: {
        placeholder: 'May 01 2020',
    },
    [FilterType.timeout]: {
        placeholder: '120s',
    },
    [FilterType.type]: {
        placeholder: 'diff|commit',
    },
    [FilterType.repohascommitafter]: {
        placeholder: 'last week',
    },
    [FilterType.after]: {
        placeholder: '"string time"',
    },
}

export interface SearchTypeLinksProps
    extends Omit<PatternTypeProps, 'setPatternType'>,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    query: string
    navbarSearchQueryState: QueryState
    onNavbarQueryChange: (queryState: QueryState) => void
}

interface SearchTypeLinkProps extends SearchTypeLinksProps {
    filterType: FilterType
    placeholder?: string
    value?: string
    children?: ReactElement | string
    onClick: (filter: FilterType) => void
}

const SearchReferenceEntry: React.FunctionComponent<SearchTypeLinkProps> = ({
    filterType,
    placeholder,
    value,
    children,
    onClick,
}) => {
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
                <button className="btn p-0" type="button" onClick={() => onClick(filterType)}>
                    <span className="text-monospace">
                        <span className="search-filter-keyword">{filterType}:</span>
                        {value ? value : <span className={styles.placeholder}>{placeholder}</span>}
                    </span>
                </button>
            </span>
            <Collapse isOpen={!collapsed}>
                <div className={styles.description}>{children}</div>
            </Collapse>
        </li>
    )
}

export const SearchReference = (props: SearchTypeLinksProps): ReactElement => {
    const filterTypes = filterTypeKeys.filter(type => type !== FilterType.patterntype)
    const [searchInput, setSearchInput] = useState('')
    const [selectedFilters, setSelectedFilters] = useState<FilterType[]>([])

    const [nextSearchValue] = useEventObservable<string, string>(
        useCallback(
            input =>
                input.pipe(
                    tap(input => setSearchInput(input)),
                    debounceTime(150),
                    distinctUntilChanged(),
                    map(input => input.trim()),
                    tap(searchValue => {
                        setSelectedFilters(
                            searchValue === ''
                                ? []
                                : filterTypeKeys.filter(filterType => filterType.indexOf(searchValue) > -1)
                        )
                    })
                ),
            [setSearchInput]
        )
    )

    const updateQuery = useCallback(
        (filterType: FilterType) => {
            // TODO: Do not just blindly append the filter. Possibly parse the
            // query and reuse existing filters if possible/necessary.
            const newQueryState: QueryState = {
                query: `${props.navbarSearchQueryState.query} ${filterType}:`,
                changeSource: QueryChangeSource.searchReference,
            }
            if (FILTERS[filterType].discreteValues) {
                newQueryState.showSuggestions = true
            } else {
                let placeholder = 'p'
                const selectionStartPosition = newQueryState.query.length + 1
                newQueryState.query += placeholder
                newQueryState.selection = new Selection(1, selectionStartPosition, 1, newQueryState.query.length + 1)
            }
            props.onNavbarQueryChange(newQueryState)
        },
        [props.onNavbarQueryChange, props.navbarSearchQueryState]
    )

    let body
    if (selectedFilters.length > 0) {
        body = (
            <ul className={styles.list}>
                {selectedFilters.map(filterType => {
                    const description = FILTERS[filterType].description
                    return (
                        <SearchReferenceEntry
                            {...props}
                            filterType={filterType}
                            key={filterType}
                            placeholder={searchReferenceInfo[filterType]?.placeholder ?? 'p'}
                            onClick={updateQuery}
                        >
                            {typeof description === 'function' ? description(false) : description}
                        </SearchReferenceEntry>
                    )
                })}
            </ul>
        )
    } else {
        body = (
            <>
                <small className={styles.header}>Match types</small>
                <ul className={styles.list}>
                    {FILTERS[FilterType.patterntype].discreteValues?.(undefined).map(({ label }) => (
                        // TODO: Use dedicated state change function to set global
                        // patterntype
                        <SearchReferenceEntry
                            {...props}
                            filterType={FilterType.patterntype}
                            key={FilterType.patterntype + label}
                            value={label}
                            onClick={updateQuery}
                        >
                            {FILTERS[FilterType.patterntype].description}
                        </SearchReferenceEntry>
                    ))}
                </ul>
                <small className={styles.header}>All Filters</small>
                <ul className={styles.list}>
                    {filterTypes.map(filterType => {
                        const description = FILTERS[filterType].description
                        return (
                            <SearchReferenceEntry
                                {...props}
                                filterType={filterType}
                                key={filterType}
                                placeholder={searchReferenceInfo[filterType]?.placeholder ?? 'p'}
                                onClick={updateQuery}
                            >
                                {typeof description === 'function' ? description(false) : description}
                            </SearchReferenceEntry>
                        )
                    })}
                </ul>
            </>
        )
    }

    return (
        <>
            <Form className={styles.filterForm} onSubmit={event => event.preventDefault()}>
                <input
                    className="form-control"
                    onChange={event => nextSearchValue(event.target.value)}
                    value={searchInput}
                    placeholder="Filter..."
                />
            </Form>
            {body}
        </>
    )
}
