import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, {
    useCallback,
    useRef,
    useEffect,
    KeyboardEvent as ReactKeyboardEvent,
    FormEvent,
    useMemo,
    useState,
} from 'react'
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
    selectSearchContextSpec: (spec: string) => void
    searchFilter: string
}> = ({ spec, description, selected, isDefault, selectSearchContextSpec, searchFilter }) => {
    const setContext = useCallback(() => {
        selectSearchContextSpec(spec)
    }, [selectSearchContextSpec, spec])
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

export interface SearchContextMenuProps
    extends Omit<SearchContextProps, 'showSearchContext' | 'setSelectedSearchContextSpec'> {
    closeMenu: () => void
    selectSearchContextSpec: (spec: string) => void
}

const getFirstMenuItem = (): HTMLButtonElement | null =>
    document.querySelector('.search-context-menu__item:first-child')

export const SearchContextMenu: React.FunctionComponent<SearchContextMenuProps> = ({
    availableSearchContexts,
    selectedSearchContextSpec,
    defaultSearchContextSpec,
    selectSearchContextSpec,
    closeMenu,
}) => {
    const inputElement = useRef<HTMLInputElement | null>(null)

    const reset = useCallback(() => {
        selectSearchContextSpec(defaultSearchContextSpec)
        closeMenu()
    }, [closeMenu, defaultSearchContextSpec, selectSearchContextSpec])

    const focusInputElement = (): void => {
        // Focus the input in the next run-loop to override the browser focusing the first dropdown item
        // if the user opened the dropdown using a keyboard
        setTimeout(() => inputElement.current?.focus(), 0)
    }

    // Reactstrap is preventing default behavior on all non-DropdownItem elements inside a Dropdown,
    // so we need to stop propagation to allow normal behavior (e.g. enter and space to activate buttons)
    // See Reactstrap bug: https://github.com/reactstrap/reactstrap/issues/2099
    const onResetButtonKeyDown = useCallback((event: ReactKeyboardEvent<HTMLButtonElement>): void => {
        if (event.key === ' ' || event.key === 'Enter') {
            event.stopPropagation()
        }
    }, [])

    useEffect(() => {
        focusInputElement()
        const onInputKeyDown = (event: KeyboardEvent): void => {
            if (event.key === 'ArrowDown') {
                getFirstMenuItem()?.focus()
                event.stopPropagation()
            }
        }
        const currentInput = inputElement.current
        currentInput?.addEventListener('keydown', onInputKeyDown)
        return () => currentInput?.removeEventListener('keydown', onInputKeyDown)
    }, [])

    useEffect(() => {
        const firstMenuItem = getFirstMenuItem()
        const onFirstMenuItemKeyDown = (event: KeyboardEvent): void => {
            if (event.key === 'ArrowUp') {
                focusInputElement()
                event.stopPropagation()
            }
        }
        firstMenuItem?.addEventListener('keydown', onFirstMenuItemKeyDown)
        return () => firstMenuItem?.removeEventListener('keydown', onFirstMenuItemKeyDown)
    }, [])

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

    const onMenuKeyDown = useCallback(
        (event: ReactKeyboardEvent<HTMLDivElement>): void => {
            if (event.key === 'Escape') {
                closeMenu()
                event.stopPropagation()
            }
        },
        [closeMenu]
    )

    return (
        <div className="search-context-menu" onKeyDown={onMenuKeyDown}>
            <div className="search-context-menu__header d-flex">
                <span aria-hidden="true" className="search-context-menu__header-prompt">
                    <ChevronRightIcon className="icon-inline" />
                </span>
                <input
                    ref={inputElement}
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
                        selectSearchContextSpec={selectSearchContextSpec}
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
                    onKeyDown={onResetButtonKeyDown}
                    className="btn btn-link btn-sm search-context-menu__footer-button"
                >
                    Reset
                </button>
                <span className="flex-grow-1" />
            </div>
        </div>
    )
}
