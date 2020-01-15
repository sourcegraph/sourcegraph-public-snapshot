import * as React from 'react'
import * as H from 'history'
import { SearchResultTabHeader } from './SearchResultTab'
import { PatternTypeProps } from '..'
import { FiltersToTypeAndValue } from '../../../../shared/src/search/interactive/util'

interface Props extends Omit<PatternTypeProps, 'setPatternType'> {
    location: H.Location
    history: H.History
    query: string
    filtersInQuery: FiltersToTypeAndValue
}

export const SearchResultTypeTabs: React.FunctionComponent<Props> = props => (
    <div className="search-result-type-tabs e2e-search-result-type-tabs border-bottom">
        <ul className="nav nav-tabs border-bottom-0">
            <SearchResultTabHeader {...props} type={null} />
            <SearchResultTabHeader {...props} type="diff" />
            <SearchResultTabHeader {...props} type="commit" />
            <SearchResultTabHeader {...props} type="symbol" />
            <SearchResultTabHeader {...props} type="repo" />
            <SearchResultTabHeader {...props} type="path" />
        </ul>
    </div>
)
