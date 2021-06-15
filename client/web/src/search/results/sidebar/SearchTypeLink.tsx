import React, { ReactElement } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '../..'
import { toggleSearchType } from '../../helpers'
import { SearchType } from '../StreamingSearchResults'

import styles from './SearchSidebarSection.module.scss'

export interface SearchTypeLinksProps
    extends Omit<PatternTypeProps, 'setPatternType'>,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    query: string
}

interface SearchTypeLinkProps extends SearchTypeLinksProps {
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

export const getSearchTypeLinks = (props: SearchTypeLinksProps): ReactElement[] => {
    const types: Exclude<SearchType, null>[] = ['file', 'path', 'symbol', 'repo', 'diff', 'commit']
    return types.map(type => <SearchTypeLink {...props} type={type} key={type} />)
}
