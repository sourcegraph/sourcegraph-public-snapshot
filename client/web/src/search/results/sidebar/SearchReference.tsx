import React, { ReactElement, useCallback, useMemo, useState } from 'react'
import classNames from 'classnames'
import { Collapse } from 'reactstrap'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../../..'
import { QueryChangeSource, QueryState, toggleSearchType } from '../../../helpers'
import { SearchType } from '../StreamingSearchResults'

import styles from './SearchReference.module.scss'
import sidebarStyles from './SearchSidebarSection.module.scss'
import {
    FilterDefinition,
    FILTERS,
    FilterType,
    filterTypeKeys,
    NegatableFilter,
} from '@sourcegraph/shared/src/search/query/filters'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { Selection } from 'monaco-editor'

interface SearchReferenceInfo {
    placeholder: string
}

const searchReferenceInfo: Record<Exclude<FilterType, NegatableFilter>, SearchReferenceInfo> &
    Record<NegatableFilter, SearchReferenceInfo> = {
    [FilterType.after]: {
        placeholder: '"string time"',
    },
    [FilterType.befoer]: {
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
            <div
                className={classNames(styles.item, sidebarStyles.sidebarSectionListItem, {
                    [styles.active]: !collapsed,
                })}
            >
                <span>
                    <button
                        type="button"
                        className={classNames('btn btn-sm', styles.collapseButton)}
                        onClick={event => {
                            event.stopPropagation()
                            setCollapsed(collapsed => !collapsed)
                        }}
                        aria-label={collapsed ? 'Expand' : 'Collapse'}
                    >
                        <CollapseIcon className="icon-inline mr-1" />
                    </button>
                    <button className="btn p-0" type="button" onClick={() => onClick(filterType)}>
                        <span className="text-code">
                            <span className="search-filter-keyword">{filterType}:</span>
                            <span className={styles.placeholder}>{value || placeholder}</span>
                        </span>
                    </button>
                </span>
            </div>
            <Collapse isOpen={!collapsed}>
                <div className={styles.description}>{children}</div>
            </Collapse>
        </li>
    )
}

export const SearchReference = (props: SearchTypeLinksProps): ReactElement => {
    const filterTypes = filterTypeKeys.filter(type => type !== FilterType.patterntype)

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

    return (
        <>
            <small>Match types</small>
            {FILTERS[FilterType.patterntype].discreteValues?.(undefined).map(({ label }) => (
                // TODO: Use dedicated state change function to set global
                // patterntype
                <ul className={sidebarStyles.sidebarSectionList}>
                    <SearchReferenceEntry
                        {...props}
                        filterType={FilterType.patterntype}
                        key={FilterType.patterntype}
                        value={label}
                        onClick={updateQuery}
                    >
                        {FILTERS[FilterType.patterntype].description}
                    </SearchReferenceEntry>
                </ul>
            ))}
            <small>All Filters</small>
            {filterTypes.map(filterType => (
                <ul className={sidebarStyles.sidebarSectionList}>
                    <SearchReferenceEntry
                        {...props}
                        filterType={filterType}
                        key={filterType}
                        placeholder={searchReferenceInfo[filterType]?.placeholder ?? 'p'}
                        onClick={updateQuery}
                    >
                        {FILTERS[filterType].description}
                    </SearchReferenceEntry>
                </ul>
            ))}
        </>
    )
}
