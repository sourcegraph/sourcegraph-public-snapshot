import React, { ReactElement } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { toggleSearchType } from '../../../helpers'
import { SearchType } from '../../SearchResults'

import { SearchSidebarProps } from './SearchSidebar'
import styles from './SearchSidebarSection.module.scss'

interface SearchTypeLinkProps extends Omit<SearchSidebarProps, 'settingsCascade'> {
    type: SearchType
}

const SearchTypeLink: React.FunctionComponent<SearchTypeLinkProps> = ({
    type,
    query,
    patternType,
    caseSensitive,
    versionContext,
    selectedSearchContextSpec,
}) => {
    const typeToggledQuery = toggleSearchType(query, type)
    const builtURLQuery = buildSearchURLQuery(
        typeToggledQuery,
        patternType,
        caseSensitive,
        versionContext,
        selectedSearchContextSpec
    )

    return (
        <Link to={{ pathname: '/search', search: builtURLQuery }} className={styles.sidebarSectionListItem}>
            <span className="text-monospace search-query-link">
                <span className="search-filter-keyword">type:</span>
                {type}
            </span>
        </Link>
    )
}

export const getSearchTypeLinks = (props: Omit<SearchSidebarProps, 'settingsCascade'>): ReactElement[] => {
    const types: Exclude<SearchType, null>[] = ['file', 'repo', 'path', 'symbol', 'diff', 'commit']
    return types.map(type => <SearchTypeLink {...props} type={type} key={type} />)
}
