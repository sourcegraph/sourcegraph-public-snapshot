import { FC, useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'

import { SearchContextInputProps, SubmitSearchProps } from '@sourcegraph/search'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Code,
    Popover,
    PopoverOpenEvent,
    Tooltip,
    PopoverContent,
    PopoverTrigger,
    Position,
    PopoverOpenEventReason,
} from '@sourcegraph/wildcard'

import { SearchContextMenu } from './SearchContextMenu'

import styles from './SearchContextDropdown.module.scss'

export interface SearchContextDropdownProps
    extends SearchContextInputProps,
        TelemetryProps,
        Partial<Pick<SubmitSearchProps, 'submitSearch'>>,
        PlatformContextProps<'requestGraphQL'> {
    query: string
    showSearchContextManagement: boolean
    authenticatedUser: AuthenticatedUser | null
    className?: string
    menuClassName?: string
    onEscapeMenuClose?: () => void
}

export const SearchContextDropdown: FC<SearchContextDropdownProps> = props => {
    const {
        authenticatedUser,
        query,
        showSearchContextManagement,
        selectedSearchContextSpec,
        setSelectedSearchContextSpec,
        fetchAutoDefinedSearchContexts,
        fetchSearchContexts,
        submitSearch,
        className,
        menuClassName,
        telemetryService,
        onEscapeMenuClose,
    } = props

    const [isOpen, setIsOpen] = useState(false)

    const selectSearchContextSpec = useCallback(
        (spec: string): void => {
            if (submitSearch) {
                submitSearch({
                    source: 'filter',
                    selectedSearchContextSpec: spec,
                })
            } else {
                setSelectedSearchContextSpec(spec)
            }
        },
        [submitSearch, setSelectedSearchContextSpec]
    )

    const handleMenuClose = useCallback(
        (isEscapeKey?: boolean) => {
            if (isEscapeKey) {
                onEscapeMenuClose?.()
            }

            setIsOpen(false)
            telemetryService.log('SearchContextDropdownToggled')
        },
        [onEscapeMenuClose, telemetryService]
    )

    const handlePopoverToggle = (event: PopoverOpenEvent): void => {
        const { isOpen, reason } = event

        setIsOpen(isOpen)

        if (reason === PopoverOpenEventReason.Esc) {
            // In order to wait until Popover will be unmounted and popover internal focus
            // management returns focus to the target, and then we call onEscapeMenuClose
            // and put focus to the search box
            requestAnimationFrame(() => onEscapeMenuClose?.())
        }

        telemetryService.log('SearchContextDropdownToggled')

        if (isOpen && authenticatedUser) {
            // Log search context dropdown view event whenever dropdown is opened, if user is authenticated
            telemetryService.log('SearchContextsDropdownViewed')
        }

        if (isOpen && !authenticatedUser) {
            // Log CTA view event whenever dropdown is opened, if user is not authenticated
            telemetryService.log('SearchResultContextsCTAShown')
        }
    }

    const isContextFilterInQuery = useMemo(() => filterExists(query, FilterType.context), [query])
    const disabledTooltipText = isContextFilterInQuery ? 'Overridden by query' : ''

    return (
        <Popover isOpen={isOpen} onOpenChange={handlePopoverToggle}>
            <Tooltip content={disabledTooltipText}>
                <PopoverTrigger
                    type="button"
                    data-testid="dropdown-toggle"
                    data-test-tooltip-content={disabledTooltipText}
                    disabled={isContextFilterInQuery}
                    className={classNames(
                        styles.button,
                        className,
                        'dropdown-toggle',
                        'test-search-context-dropdown',
                        isOpen && styles.buttonOpen
                    )}
                >
                    <Code className={classNames('test-selected-search-context-spec', styles.buttonContent)}>
                        {
                            // a11y-ignore
                            // Rule: "color-contrast" (Elements must have sufficient color contrast)
                            // GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/33343
                        }
                        <span className="a11y-ignore search-filter-keyword">context</span>
                        <span className="search-filter-separator">:</span>
                        {selectedSearchContextSpec?.startsWith('@') ? (
                            <>
                                <span className="search-keyword">@</span>
                                {selectedSearchContextSpec?.slice(1)}
                            </>
                        ) : (
                            selectedSearchContextSpec
                        )}
                    </Code>
                </PopoverTrigger>
            </Tooltip>
            {/*
               a11y-ignore
               Rule: "aria-required-children" (Certain ARIA roles must contain particular children)
               GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/34348
             */}
            <PopoverContent
                position={Position.bottomStart}
                className={classNames('a11y-ignore', styles.menu)}
                data-testid="dropdown-content"
            >
                <SearchContextMenu
                    {...props}
                    showSearchContextManagement={showSearchContextManagement}
                    selectSearchContextSpec={selectSearchContextSpec}
                    fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                    fetchSearchContexts={fetchSearchContexts}
                    className={menuClassName}
                    onMenuClose={handleMenuClose}
                />
            </PopoverContent>
        </Popover>
    )
}
