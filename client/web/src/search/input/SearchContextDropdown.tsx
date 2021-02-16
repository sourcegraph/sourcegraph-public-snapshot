import classNames from 'classnames'
import React, { useCallback, useEffect, useState } from 'react'
import { Dropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import { isContextFilterInQuery, SearchContextProps } from '..'
import { SearchContextMenu } from './SearchContextMenu'

export interface SearchContextDropdownProps extends Omit<SearchContextProps, 'showSearchContext'> {
    query: string
}

export const SearchContextDropdown: React.FunctionComponent<SearchContextDropdownProps> = props => {
    const { query, selectedSearchContextSpec } = props

    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(value => !value), [])
    const [isDisabled, setIsDisabled] = useState(false)

    // Disable the dropdown if the query contains a context filter
    useEffect(() => setIsDisabled(isContextFilterInQuery(query)), [query])

    return (
        <>
            <Dropdown
                isOpen={isOpen}
                toggle={toggleOpen}
                a11y={false} /* Override default keyboard events in reactstrap */
            >
                <DropdownToggle
                    className={classNames('search-context-dropdown__button', 'dropdown-toggle', {
                        'search-context-dropdown__button--open': isOpen,
                    })}
                    color="link"
                    disabled={isDisabled}
                    data-tooltip={isDisabled ? 'Overridden by query' : ''}
                >
                    <code className="search-context-dropdown__button-content">
                        <span className="search-filter-keyword">context:</span>
                        {selectedSearchContextSpec.startsWith('@') ? (
                            <>
                                <span className="search-keyword">@</span>
                                {selectedSearchContextSpec.slice(1)}
                            </>
                        ) : (
                            selectedSearchContextSpec
                        )}
                    </code>
                </DropdownToggle>
                <DropdownMenu>
                    <SearchContextMenu {...props} closeMenu={toggleOpen} />
                </DropdownMenu>
            </Dropdown>
            <div className="search-context-dropdown__separator" />
        </>
    )
}
