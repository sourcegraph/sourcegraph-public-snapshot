import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import { SearchContextMenu } from './SearchContextMenu'

interface SearchContextDropdownProps {
    searchContextSpec: string
}

export const SearchContextDropdown: React.FunctionComponent<SearchContextDropdownProps> = ({ searchContextSpec }) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(value => !value), [])

    return (
        <>
            <ButtonDropdown isOpen={isOpen} toggle={toggleOpen}>
                <DropdownToggle className="search-context-dropdown__button dropdown-toggle" color="link">
                    <code className="search-context-dropdown__button-content">
                        <span className="search-filter-keyword">context:</span>
                        {searchContextSpec}
                    </code>
                </DropdownToggle>
                <DropdownMenu>
                    <SearchContextMenu />
                </DropdownMenu>
            </ButtonDropdown>
            <div className="search-context-dropdown__separator" />
        </>
    )
}
