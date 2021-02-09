import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import { SearchContextMenu } from './SearchContextMenu'

interface SearchContextDropdownProps {}

export const SearchContextDropdown: React.FunctionComponent<SearchContextDropdownProps> = () => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(value => !value), [])

    const context = 'global'
    return (
        <>
            <ButtonDropdown isOpen={isOpen} toggle={toggleOpen}>
                <DropdownToggle className="search-context-dropdown__button dropdown-toggle" color="link">
                    <code className="search-context-dropdown__button-content">
                        <span className="search-filter-keyword">context:</span>
                        {context}
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
