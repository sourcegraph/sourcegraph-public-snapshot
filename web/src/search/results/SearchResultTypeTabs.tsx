import * as React from 'react'
import SearchResultTab from './SearchResultTab'
import { SEARCH_TYPES } from './SearchResults'

interface Props {
    activeType: SEARCH_TYPES
    query: string
    onTabClicked: (query: SEARCH_TYPES) => void
}

export default class SearchResultTypeTabs extends React.Component<Props> {
    constructor(props: Props) {
        super(props)
    }

    public render(): JSX.Element | null {
        return (
            <div className="search-result-type-tabs">
                <SearchResultTab
                    active={this.props.activeType === ''}
                    type=""
                    onClick={this.props.onTabClicked}
                    query={this.props.query}
                />
                <SearchResultTab
                    active={this.props.activeType === 'diff'}
                    type="diff"
                    onClick={this.props.onTabClicked}
                    query={this.props.query}
                />
                <SearchResultTab
                    active={this.props.activeType === 'commit'}
                    type="commit"
                    onClick={this.props.onTabClicked}
                    query={this.props.query}
                />
                <SearchResultTab
                    active={this.props.activeType === 'symbol'}
                    type="symbol"
                    onClick={this.props.onTabClicked}
                    query={this.props.query}
                />
                <SearchResultTab
                    active={this.props.activeType === 'repo'}
                    type="repo"
                    onClick={this.props.onTabClicked}
                    query={this.props.query}
                />
            </div>
        )
    }
}
