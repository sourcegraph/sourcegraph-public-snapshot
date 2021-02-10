import classNames from 'classnames'
import React, { useCallback, useEffect, useState } from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import { scanSearchQuery } from '../../../../shared/src/search/query/scanner'
import { FilterType } from '../../../../shared/src/search/query/filters'
import { SearchContextMenu } from './SearchContextMenu'

interface SearchContextDropdownProps {
    query: string
}

export const SearchContextDropdown: React.FunctionComponent<SearchContextDropdownProps> = ({ query }) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => setIsOpen(value => !value), [])
    const [isDisabled, setIsDisabled] = useState(false)

    // Disable the dropdown if the query contains a context filter
    useEffect(() => {
        const scannedQuery = scanSearchQuery(query)
        const isDisabled =
            scannedQuery.type === 'success' &&
            scannedQuery.term.some(
                token => token.type === 'filter' && token.field.value.toLowerCase() === FilterType.context
            )
        setIsDisabled(isDisabled)
    }, [query])

    const context = 'global'
    return (
        <>
            <ButtonDropdown isOpen={isOpen} toggle={toggleOpen}>
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
