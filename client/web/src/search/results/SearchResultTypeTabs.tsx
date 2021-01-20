import * as H from 'history'
import * as React from 'react'
import classNames from 'classnames'
import { CaseSensitivityProps, InteractiveSearchProps, PatternTypeProps } from '..'
import { SearchResultTabHeader } from './SearchResultTab'
import { VersionContextProps } from '../../../../shared/src/search/util'

interface Props
    extends Omit<PatternTypeProps, 'setPatternType'>,
        Omit<CaseSensitivityProps, 'setCaseSensitivity'>,
        Partial<Pick<InteractiveSearchProps, 'filtersInQuery'>>,
        VersionContextProps {
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
