import React from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'

interface SearchContextDropdownProps {}

export const SearchContextDropdown: React.FunctionComponent<SearchContextDropdownProps> = () => {
    const context = 'global'
    return (
        <>
            <ButtonDropdown>
                <DropdownToggle className="btn btn-link text-monospace search-context-dropdown__button">
                    <span className="search-filter-keyword">context:</span>
                    {context}
                </DropdownToggle>
            </ButtonDropdown>
            <div className="search-context-dropdown__separator" />
        </>
    )
}
