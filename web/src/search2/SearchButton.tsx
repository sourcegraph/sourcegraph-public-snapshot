import SearchIcon from '@sourcegraph/icons/lib/Search'
import * as React from 'react'

export class SearchButton extends React.Component {
    public render(): JSX.Element | null {
        return (
            <button className="search-button2 btn btn-primary" type="submit">
                <SearchIcon className="icon-inline" />
                <span className="search-button2__label">Search code</span>
            </button>
        )
    }
}
