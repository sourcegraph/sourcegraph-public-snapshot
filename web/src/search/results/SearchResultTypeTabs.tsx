import * as React from 'react'
import * as H from 'history'
import { SearchResultTabHeader } from './SearchResultTab'

interface Props {
    location: H.Location
    history: H.History
    query: string
}
export const SearchResultTypeTabs: React.FunctionComponent<Props> = props => (
    <div className="search-result-type-tabs e2e-search-result-type-tabs border-bottom">
        <ul className="nav nav-tabs border-bottom-0">
            <SearchResultTabHeader {...props} type={null} />
            <SearchResultTabHeader {...props} type="diff" />
            <SearchResultTabHeader {...props} type="commit" />
            <SearchResultTabHeader {...props} type="symbol" />
            <SearchResultTabHeader {...props} type="repo" />
        </ul>
    </div>
)
