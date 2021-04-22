import { Link } from '@sourcegraph/shared/src/components/Link'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import React from 'react'
import { toggleSearchType } from '../../../helpers'
import { SearchType } from '../../SearchResults'
import { SearchSidebarProps } from './SearchSidebar'
import styles from './SearchSidebar.module.scss'

interface ResultTypeLinkProps extends SearchSidebarProps {
    type: SearchType
}

const ResultTypeLink: React.FunctionComponent<ResultTypeLinkProps> = ({
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

export const SidebarResultTypeSection: React.FunctionComponent<SearchSidebarProps> = props => {
    const types: SearchType[] = ['file', 'repo', 'path', 'symbol', 'diff', 'commit']

    return (
        <ul className={styles.sidebarSectionList}>
            {types.map(type => (
                <li>
                    <ResultTypeLink {...props} type={type} />
                </li>
            ))}
        </ul>
    )
}
