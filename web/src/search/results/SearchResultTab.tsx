import * as React from 'react'
import * as H from 'history'
import { SEARCH_TYPES } from './SearchResults'
import { NavLink } from 'react-router-dom'
import { toggleSearchType } from '../helpers'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'

interface Props {
    location: H.Location
    history: H.History
    type: SEARCH_TYPES
    query: string
}

const typeToProse: Record<SEARCH_TYPES, string> = {
    '': 'Code',
    diff: 'Diffs',
    commit: 'Commits',
    symbol: 'Symbols',
    repo: 'Repos',
}

const tabIsActive = (builtQuery: string, location: H.Location): boolean => location.search === '?' + builtQuery

const tabIsActiveTrue = (): boolean => true
const tabIsActiveFalse = (): boolean => false

export const SearchResultTab: React.FunctionComponent<Props> = props => {
    const q = toggleSearchType(props.query, props.type)
    const newURLSearchParam = buildSearchURLQuery(q)

    const isActiveFunc = tabIsActive(newURLSearchParam, props.location) ? tabIsActiveTrue : tabIsActiveFalse
    return (
        <li className="nav-item">
            <NavLink
                to={{ pathname: '/search', search: newURLSearchParam }}
                className="nav-link"
                activeClassName="active"
                isActive={isActiveFunc}
            >
                {typeToProse[props.type]}
            </NavLink>
        </li>
    )
}
