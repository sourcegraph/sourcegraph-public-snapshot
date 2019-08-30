import * as React from 'react'
import SearchResultTab from './SearchResultTab'
import { SEARCH_TYPES } from './SearchResults'

interface Props {
    activeType: SEARCH_TYPES
    query: string
    onTabClicked: (query: SEARCH_TYPES) => void
}

export const SearchResultTypeTabs: React.FunctionComponent<{
    activeType: SEARCH_TYPES
    query: string
    onTabClicked: (query: SEARCH_TYPES) => void
}> = ({ activeType, query, onTabClicked }) => (
    <div className="search-result-type-tabs">
        <SearchResultTab active={activeType === ''} type="" onClick={onTabClicked} query={query} />
        <SearchResultTab active={activeType === 'diff'} type="diff" onClick={onTabClicked} query={query} />
        <SearchResultTab active={activeType === 'commit'} type="commit" onClick={onTabClicked} query={query} />
        <SearchResultTab active={activeType === 'symbol'} type="symbol" onClick={onTabClicked} query={query} />
        <SearchResultTab active={activeType === 'repo'} type="repo" onClick={onTabClicked} query={query} />
    </div>
)
