import { Link } from '@sourcegraph/shared/src/components/Link'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import React from 'react'
import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../../..'
import { toggleSearchType } from '../../../helpers'
import { SearchType } from '../../SearchResults'
import styles from './SearchSidebar.module.scss'

interface ResultTypeLinkProps
    extends Omit<PatternTypeProps, 'setPatternType'>,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    type: SearchType
    query: string
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

export const SidebarResultTypeSection: React.FunctionComponent<Omit<ResultTypeLinkProps, 'type'>> = props => {
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
