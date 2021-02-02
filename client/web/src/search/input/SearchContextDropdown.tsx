import React from 'react'

interface SearchContextDropdownProps {}

export const SearchContextDropdown: React.FunctionComponent<SearchContextDropdownProps> = () => {
    const context = 'global'
    return (
        <>
            <button type="button" className="btn btn-link text-monospace search-context-dropdown__button">
                <span className="search-filter-keyword">context:</span>
                {context}
            </button>
            <div className="search-context-dropdown__separator" />
        </>
    )
}
