import * as React from 'react'
import { SEARCH_TYPES } from './SearchResults'

interface Props {
    type: SEARCH_TYPES
    onClick: (query: SEARCH_TYPES) => void
    query: string
}

export default class SearchResultTab extends React.Component<Props, {}> {
    constructor(props: Props) {
        super(props)
    }

    public render(): JSX.Element | null {
        return (
            <div className="search-result-tab">
                <button className="btn" onClick={this.onClick}>
                    {this.props.type}
                </button>
            </div>
        )
    }

    private onClick: React.MouseEventHandler<HTMLButtonElement> = event => {
        event.preventDefault()
        this.props.onClick(this.props.type)
    }
}
