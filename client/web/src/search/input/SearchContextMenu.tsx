import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { FormEvent, useCallback, useMemo, useState } from 'react'
import { DropdownItem } from 'reactstrap'
import { SearchContextProps } from '..'

const HighlightedSearchTerm: React.FunctionComponent<{ text: string; searchFilter: string }> = ({
    text,
    searchFilter,
}) => {
    if (searchFilter.length > 0) {
        const index = text.toLowerCase().indexOf(searchFilter.toLowerCase())
        if (index > -1) {
            const before = text.slice(0, index)
            const highlighted = text.slice(index, index + searchFilter.length)
            const after = text.slice(index + searchFilter.length)
            return (
                <>
                    {before}
                    <strong>{highlighted}</strong>
                    {after}
                </>
            )
        }
    }
    return <>{text}</>
}

const SearchContextMenuItem: React.FunctionComponent<{
    spec: string
    description: string
    selected: boolean
    isDefault: boolean
    setSelectedSearchContextSpec: (spec: string) => void
    searchFilter: string
}> = ({ spec, description, selected, isDefault, setSelectedSearchContextSpec, searchFilter }) => {
    const setContext = useCallback(() => {
        setSelectedSearchContextSpec(spec)
    }, [setSelectedSearchContextSpec, spec])
    return (
        <DropdownItem
            className={classNames('search-context-menu__item', { 'search-context-menu__item--selected': selected })}
            onClick={setContext}
        >
            <span className="search-context-menu__item-name" title={spec}>
                <HighlightedSearchTerm text={spec} searchFilter={searchFilter} />
            </span>{' '}
            <span className="search-context-menu__item-description" title={description}>
                {description}
            </span>
            {isDefault && <span className="search-context-menu__item-default">Default</span>}
        </DropdownItem>
    )
}

export interface SearchContextMenuProps extends Omit<SearchContextProps, 'showSearchContext'> {
    closeMenu: () => void
}

export const SearchContextMenu: React.FunctionComponent<SearchContextMenuProps> = ({
    availableSearchContexts,
    selectedSearchContextSpec,
    defaultSearchContextSpec,
    setSelectedSearchContextSpec,
    closeMenu,
}) => {
    const reset = useCallback(() => {
        setSelectedSearchContextSpec(defaultSearchContextSpec)
        closeMenu()
    }, [closeMenu, defaultSearchContextSpec, setSelectedSearchContextSpec])

    const [searchFilter, setSearchFilter] = useState('')
    const onSearchFilterChanged = useCallback(
        (event: FormEvent<HTMLInputElement>) => setSearchFilter(event ? event.currentTarget.value : ''),
        []
    )

    const filteredList = useMemo(
        () =>
            availableSearchContexts.filter(context => context.spec.toLowerCase().includes(searchFilter.toLowerCase())),
        [availableSearchContexts, searchFilter]
    )

    return (
        <div className="search-context-menu">
            <div className="search-context-menu__header d-flex">
                <span aria-hidden="true" className="search-context-menu__header-prompt">
                    <ChevronRightIcon className="icon-inline" />
                </span>
                <input
                    onInput={onSearchFilterChanged}
                    type="search"
                    placeholder="Find a context"
                    className="search-context-menu__header-input"
                />
            </div>
            <div className="search-context-menu__list">
                {filteredList.map(context => (
                    <SearchContextMenuItem
                        key={context.id}
                        spec={context.spec}
                        description={context.description}
                        isDefault={context.spec === defaultSearchContextSpec}
                        selected={context.spec === selectedSearchContextSpec}
                        setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                        searchFilter={searchFilter}
                    />
                ))}
                {filteredList.length === 0 && (
                    <DropdownItem className="search-context-menu__item" disabled={true}>
                        No contexts found
                    </DropdownItem>
                )}
            </div>
            <div className="search-context-menu__footer">
                <button
                    type="button"
                    onClick={reset}
                    className="btn btn-link btn-sm search-context-menu__footer-button"
                >
                    Reset
                </button>
                <span className="flex-grow-1" />
            </div>
        </div>
    )
}
