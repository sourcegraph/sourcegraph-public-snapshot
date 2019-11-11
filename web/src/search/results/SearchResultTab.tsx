import * as React from 'react'
import * as H from 'history'
import { SearchType } from './SearchResults'
import { NavLink } from 'react-router-dom'
import { toggleSearchType } from '../helpers'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { constant } from 'lodash'
import { PatternTypeProps } from '..'

interface Props extends Omit<PatternTypeProps, 'togglePatternType'> {
    location: H.Location
    type: SearchType
    query: string
}

const typeToProse: Record<Exclude<SearchType, null>, string> = {
    diff: 'Diffs',
    commit: 'Commits',
    symbol: 'Symbols',
    repo: 'Repos',
    path: 'Files',
}

export const SearchResultTabHeader: React.FunctionComponent<Props> = ({ location, type, query, patternType }) => {
    const q = toggleSearchType(query, type)
    const builtURLQuery = buildSearchURLQuery(q, patternType)

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
