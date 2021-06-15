import React, { ReactElement, useState } from 'react'
import classNames from 'classnames'
import { Collapse } from 'reactstrap'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../../..'
import { QueryState, toggleSearchType } from '../../../helpers'
import { SearchType } from '../StreamingSearchResults'

import styles from './SearchReference.module.scss'
import sidebarStyles from './SearchSidebarSection.module.scss'
import { FilterDefinition, FILTERS, FilterType, filterTypeKeys } from '@sourcegraph/shared/src/search/query/filters'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'

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
        <div
            className={classNames(styles.searchReferenceItem, sidebarStyles.sidebarSectionListItem)}
            onClick={() => onClick(filterType)}
        >
            <span className="text-monospace">
                <button
                    type="button"
                    className={classNames('btn btn-', styles.searchReferenceCollapseButton)}
                    onClick={() => setCollapsed(collapsed => !collapsed)}
                    aria-label={collapsed ? 'Expand' : 'Collapse'}
                >
                    <CollapseIcon className="icon-inline mr-1" />
                </button>
                <span className="search-filter-keyword">{filterType}:</span>
                <span className={styles.searchReferencePlaceholder}>{value || placeholder}</span>
            </span>
            <Collapse isOpen={!collapsed}>
                <div className={styles.searchReferenceDescription}>{children}</div>
            </Collapse>
        </div>
    )
}

export const SearchReference = (props: SearchTypeLinksProps): ReactElement => {
    const filterTypes = filterTypeKeys.filter(type => type !== FilterType.patterntype)
    return (
        <>
            <div className={styles.searchReferenceHeader}>Match types</div>
            {FILTERS[FilterType.patterntype].discreteValues?.(undefined).map(({ label }) => (
                <SearchReferenceEntry
                    {...props}
                    filterType={FilterType.patterntype}
                    key={FilterType.patterntype}
                    value={label}
                    onClick={(filterType: FilterType) =>
                        props.onNavbarQueryChange({
                            ...props.navbarSearchQueryState,
                            query: `${props.navbarSearchQueryState.query} ${filterType}:`,
                        })
                    }
                >
                    {FILTERS[FilterType.patterntype].description}
                </SearchReferenceEntry>
            ))}
            <div className={styles.searchReferenceHeader}>All Filters</div>
            {filterTypes.map(filterType => (
                <SearchReferenceEntry
                    {...props}
                    filterType={filterType}
                    key={filterType}
                    placeholder={'p'}
                    onClick={(filterType: FilterType) =>
                        props.onNavbarQueryChange({
                            ...props.navbarSearchQueryState,
                            query: `${props.navbarSearchQueryState.query} ${filterType}:`,
                        })
                    }
                >
                    {FILTERS[filterType].description}
                </SearchReferenceEntry>
            ))}
        </>
    )
}
