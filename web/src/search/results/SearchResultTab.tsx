import * as React from 'react'
import * as H from 'history'
import { SearchType } from './SearchResults'
import { NavLink } from 'react-router-dom'
import { toggleSearchType } from '../helpers'
import { buildSearchURLQuery, generateFiltersQuery } from '../../../../shared/src/util/url'
import { constant } from 'lodash'
import { PatternTypeProps, CaseSensitivityProps } from '..'
import { FiltersToTypeAndValue } from '../../../../shared/src/search/interactive/util'

interface Props extends Omit<PatternTypeProps, 'setPatternType'>, Omit<CaseSensitivityProps, 'setCaseSensitivity'> {
    location: H.Location
    type: SearchType
    query: string
    filtersInQuery: FiltersToTypeAndValue
}

const typeToProse: Record<Exclude<SearchType, null>, string> = {
    diff: 'Diffs',
    commit: 'Commits',
    symbol: 'Symbols',
    repo: 'Repositories',
    path: 'Filenames',
}

export const SearchResultTabHeader: React.FunctionComponent<Props> = ({
    location,
    type,
    query,
    filtersInQuery,
    patternType,
    caseSensitive,
}) => {
    const fullQuery = [query, generateFiltersQuery(filtersInQuery)].filter(query => query.length > 0).join(' ')
    const q = toggleSearchType(fullQuery, type)
    const builtURLQuery = buildSearchURLQuery(q, patternType, caseSensitive)

    const isActiveFunc = constant(location.search === `?${builtURLQuery}`)
    return (
        <li className="nav-item e2e-search-result-tab">
            <NavLink
                to={{ pathname: '/search', search: builtURLQuery }}
                className={`nav-link e2e-search-result-tab-${type}`}
                activeClassName="active e2e-search-result-tab--active"
                isActive={isActiveFunc}
            >
                {type ? typeToProse[type] : 'Code'}
            </NavLink>
        </li>
    )
}
