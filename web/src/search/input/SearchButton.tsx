import SearchIcon from 'mdi-react/SearchIcon'
import React from 'react'

/**
 * A search button. It must be wrapped in a form whose onSubmit handler performs the search.
 */
export const SearchButton: React.FunctionComponent = () => (
    <div className="search-button d-flex text-nowrap">
        <button className="btn btn-primary search-button__btn" type="submit">
            <SearchIcon className="icon-inline" />
        </button>
    </div>
)
