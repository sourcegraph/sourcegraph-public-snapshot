import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React, { useCallback, useRef, useEffect } from 'react'
import { DropdownItem } from 'reactstrap'
import { SearchContextProps } from '..'

const SearchContextMenuItem: React.FunctionComponent<{
    spec: string
    description: string
    selected: boolean
    isDefault: boolean
    setSelectedSearchContextSpec: (spec: string) => void
}> = ({ spec, description, selected, isDefault, setSelectedSearchContextSpec }) => {
    const setContext = useCallback(() => {
        setSelectedSearchContextSpec(spec)
    }, [setSelectedSearchContextSpec, spec])
    return (
        <DropdownItem
            className={classNames('search-context-menu__item', { 'search-context-menu__item--selected': selected })}
            onClick={setContext}
        >
            <span className="search-context-menu__item-name" title={spec}>
                {spec}
            </span>
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

const getFirstMenuItem = (): HTMLButtonElement | null =>
    document.querySelector('.search-context-menu__item:first-child')

export const SearchContextMenu: React.FunctionComponent<SearchContextMenuProps> = ({
    availableSearchContexts,
    selectedSearchContextSpec,
    defaultSearchContextSpec,
    setSelectedSearchContextSpec,
    closeMenu,
}) => {
    const inputElement = useRef<HTMLInputElement | null>(null)

    const reset = useCallback(() => {
        setSelectedSearchContextSpec(defaultSearchContextSpec)
        closeMenu()
    }, [closeMenu, defaultSearchContextSpec, setSelectedSearchContextSpec])

    const focusInputElement = (): void => {
        // Focus the input in the next run-loop to override the browser focusing the first dropdown item
        // if the user opened the dropdown using a keyboard
        setTimeout(() => inputElement.current?.focus(), 0)
    }

    useEffect(() => {
        focusInputElement()
        const onInputKeyDown = (event: KeyboardEvent): void => {
            if (event.key === 'ArrowDown') {
                getFirstMenuItem()?.focus()
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
            }
        }
        firstMenuItem?.addEventListener('keydown', onFirstMenuItemKeyDown)
        return () => firstMenuItem?.removeEventListener('keydown', onFirstMenuItemKeyDown)
    }, [])

    return (
        <div className="search-context-menu">
            <div className="search-context-menu__header d-flex">
                <span aria-hidden="true" className="search-context-menu__header-prompt">
                    <ChevronRightIcon className="icon-inline" />
                </span>
                <input
                    ref={inputElement}
                    type="search"
                    placeholder="Find a context"
                    className="search-context-menu__header-input"
                />
            </div>
            <div className="search-context-menu__list">
                {availableSearchContexts.map(context => (
                    <SearchContextMenuItem
                        key={context.id}
                        spec={context.spec}
                        description={context.description}
                        isDefault={context.spec === defaultSearchContextSpec}
                        selected={context.spec === selectedSearchContextSpec}
                        setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                    />
                ))}
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
