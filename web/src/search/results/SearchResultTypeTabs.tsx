import * as React from 'react'
import SearchResultTab from './SearchResultTab'
import { SEARCH_TYPES } from './SearchResults'

interface Props {
    query: string
    onTabClicked: (query: SEARCH_TYPES) => void
}

export default class SearchResultTypeTabs extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        return (
            <div className="search-result-type-tabs">
                <SearchResultTab type={'code'} onClick={this.props.onTabClicked} query={this.props.query} />
                <SearchResultTab type={'diff'} onClick={this.props.onTabClicked} query={this.props.query} />
                <SearchResultTab type={'commit'} onClick={this.props.onTabClicked} query={this.props.query} />
                <SearchResultTab type={'symbol'} onClick={this.props.onTabClicked} query={this.props.query} />
            </div>
        )
    }
}
