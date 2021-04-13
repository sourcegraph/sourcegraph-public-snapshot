import classNames from 'classnames'
import * as H from 'history'
import * as React from 'react'

import { VersionContextProps } from '@sourcegraph/shared/src/search/util'

import { CaseSensitivityProps, ParsedSearchQueryProps, PatternTypeProps, SearchContextProps } from '..'

import { SearchResultTabHeader } from './SearchResultTab'

interface Props
    extends Omit<PatternTypeProps, 'setPatternType'>,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    location: H.Location
    history: H.History
    query: string
    className?: string
}

export const SearchResultTypeTabs: React.FunctionComponent<Props> = props => (
    <div className={classNames(props.className, 'mt-2 border-bottom test-search-result-type-tabs')}>
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
