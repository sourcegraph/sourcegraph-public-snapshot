import SearchIcon from 'mdi-react/SearchIcon'
import * as React from 'react'

export class SearchButton extends React.Component {
    public render(): JSX.Element | null {
        return (
            <button className="search-button btn btn-primary" type="submit">
                <SearchIcon className="icon-inline" />
                <span className="search-button__label">Search</span>
            </button>
        )
    }
}
