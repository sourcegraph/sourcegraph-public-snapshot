import * as React from 'react'
import { SEARCH_TYPES } from './SearchResults'
import { startCase } from 'lodash'

interface Props {
    active: boolean
    type: SEARCH_TYPES
    onClick: (query: SEARCH_TYPES) => void
    query: string
}

const typeToProse: Record<SEARCH_TYPES, string> = {
    '': 'Code',
    diff: 'Diffs',
    commit: 'Commits',
    symbol: 'Symbols',
    repo: 'Repos',
}

export default class SearchResultTab extends React.Component<Props, {}> {
    constructor(props: Props) {
        super(props)
    }

    public render(): JSX.Element | null {
        return (
            <button
                className={`btn search-result-tab ${this.props.active && 'search-result-tab--active'}`}
                onClick={this.onClick}
            >
                <div className="search-result-tab__inner">{typeToProse[this.props.type]}</div>
            </button>
        )
    }

    private onClick: React.MouseEventHandler<HTMLButtonElement> = event => {
        event.preventDefault()
        this.props.onClick(this.props.type)
    }
}
