import React from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'

interface SearchContextDropdownProps {}

export const SearchContextDropdown: React.FunctionComponent<SearchContextDropdownProps> = () => {
    const context = 'global'
    return (
        <>
            <ButtonDropdown>
                <DropdownToggle className="search-context-dropdown__button" color="link">
                    <code>
                        <span className="search-filter-keyword">context:</span>
                        {context}
                    </code>
                </DropdownToggle>
            </ButtonDropdown>
            <div className="search-context-dropdown__separator" />
        </>
    )
}
