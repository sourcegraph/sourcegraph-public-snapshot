import * as React from 'react'
import SearchResultTab from './SearchResultTab'

interface Props {
    query: string
}
export const SearchResultTypeTabs: React.FunctionComponent<Props> = props => (
    <div className="search-result-type-tabs e2e-search-result-type-tabs border-bottom">
        <ul className="nav nav-tabs border-bottom-0">
            <SearchResultTab type="" query={props.query} />
            <SearchResultTab type="diff" query={props.query} />
            <SearchResultTab type="commit" query={props.query} />
            <SearchResultTab type="symbol" query={props.query} />
            <SearchResultTab type="repo" query={props.query} />
        </ul>
    </div>
)
