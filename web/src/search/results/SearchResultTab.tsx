import * as React from 'react'
import { SEARCH_TYPES } from './SearchResults'
import { NavLink } from 'react-router-dom'
import { toggleSearchType } from '../helpers'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'

interface Props {
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

export default class SearchResultTab extends React.Component<Props> {
    public render(): JSX.Element | null {
        const q = toggleSearchType(this.props.query, this.props.type)
        const newURLSearchParam = buildSearchURLQuery(q)

        return (
            <li className="nav-item">
                <NavLink
                    to={{ pathname: '/search', search: newURLSearchParam }}
                    className={`nav-link`}
                    activeClassName={'active'}
                    isActive={(_, location) => location.search === '?' + newURLSearchParam}
                >
                    {typeToProse[this.props.type]}
                </NavLink>
            </li>
        )
    }
}
