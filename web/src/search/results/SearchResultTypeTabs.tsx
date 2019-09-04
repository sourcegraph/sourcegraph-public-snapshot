import * as React from 'react'
import * as H from 'history'
import { SearchResultTab } from './SearchResultTab'

interface Props {
    location: H.Location
    history: H.History
    query: string
}
export const SearchResultTypeTabs: React.FunctionComponent<Props> = props => (
    <div className="search-result-type-tabs e2e-search-result-type-tabs border-bottom">
        <ul className="nav nav-tabs border-bottom-0">
            <SearchResultTab {...props} type="" />
            <SearchResultTab {...props} type="diff" />
            <SearchResultTab {...props} type="commit" />
            <SearchResultTab {...props} type="symbol" />
            <SearchResultTab {...props} type="repo" />
        </ul>
    </div>
)
