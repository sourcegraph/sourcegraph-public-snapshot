import { constant } from 'lodash'
import * as React from 'react'
import { NavLink } from 'react-router-dom'

import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { PatternTypeProps, CaseSensitivityProps, ParsedSearchQueryProps, SearchContextProps } from '..'
import { toggleSearchType } from '../helpers'

import { SearchType } from './SearchResults'

interface Props
    extends Omit<PatternTypeProps, 'setPatternType'>,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    type: SearchType
    query: string
}

const typeToProse: Record<Exclude<SearchType, null>, string> = {
    diff: 'Diffs',
    commit: 'Commits',
    symbol: 'Symbols',
    repo: 'Repositories',
    path: 'Filenames',
    file: 'File contents',
}

export const SearchResultTabHeader: React.FunctionComponent<Props> = ({
    type,
    query,
    parsedSearchQuery,
    patternType,
    caseSensitive,
    versionContext,
    selectedSearchContextSpec,
}) => {
    const caseToggledQuery = toggleSearchType(query, type)
    const builtURLQuery = buildSearchURLQuery(
        caseToggledQuery,
        patternType,
        caseSensitive,
        versionContext,
        selectedSearchContextSpec
    )

    const currentQuery = parsedSearchQuery
    const scannedQuery = scanSearchQuery(currentQuery)
    let typeInQuery: SearchType = null

    if (scannedQuery.type === 'success') {
        // Parse any `type:` filter that exists in a query so
        // we can check whether this tab should be active.
        for (const token of scannedQuery.term) {
            if (token.type === 'filter' && token.field.value === 'type' && token.value) {
                typeInQuery = token.value.value as SearchType
            }
        }
    }

    const isActiveFunc = constant(typeInQuery === type)
    return (
        <li className="nav-item test-search-result-tab">
            <NavLink
                to={{ pathname: '/search', search: builtURLQuery }}
                className={`nav-link test-search-result-tab-${String(type)}`}
                activeClassName="active test-search-result-tab--active"
                isActive={isActiveFunc}
            >
                {type ? typeToProse[type] : 'Code'}
            </NavLink>
        </li>
    )
}
