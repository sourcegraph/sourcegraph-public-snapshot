import { type FC, useCallback, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SearchContextInputProps, SubmitSearchProps } from '@sourcegraph/shared/src/search'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Code,
    Popover,
    type PopoverOpenEvent,
    Tooltip,
    PopoverContent,
    PopoverTrigger,
    Position,
    PopoverOpenEventReason,
    useUpdateEffect,
    createRectangle,
} from '@sourcegraph/wildcard'

import { SearchContextMenu } from './SearchContextMenu'

import styles from './SearchContextDropdown.module.scss'

// Adds padding to the popover content to add some space between the trigger
// button and the content
const popoverPadding = createRectangle(0, 0, 0, 2)

export interface SearchContextDropdownProps
    extends SearchContextInputProps,
        TelemetryProps,
        Partial<Pick<SubmitSearchProps, 'submitSearch'>>,
        PlatformContextProps<'requestGraphQL'> {
    query: string
    showSearchContextManagement: boolean
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean | null
    className?: string
    menuClassName?: string
    onEscapeMenuClose?: () => void
    ignoreDefaultContextDoesNotExistError?: boolean
}

export const SearchContextDropdown: FC<SearchContextDropdownProps> = props => {
    const {
        authenticatedUser,
        query,
        showSearchContextManagement,
        selectedSearchContextSpec,
        setSelectedSearchContextSpec,
        fetchSearchContexts,
        submitSearch,
        className,
        menuClassName,
        telemetryService,
        onEscapeMenuClose,
    } = props

    const [isOpen, setIsOpen] = useState(false)
    const searchContextDropdownReference = useRef<HTMLButtonElement>(null)

    useUpdateEffect(() => {
        if (!isOpen) {
            searchContextDropdownReference?.current?.focus()
        }
    }, [isOpen])

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
                    ref={searchContextDropdownReference}
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
            <PopoverContent
                position={Position.bottomStart}
                className={classNames('a11y-ignore overflow-hidden', styles.menu)}
                data-testid="dropdown-content"
                targetPadding={popoverPadding}
            >
                <SearchContextMenu
                    {...props}
                    showSearchContextManagement={showSearchContextManagement}
                    selectSearchContextSpec={selectSearchContextSpec}
                    fetchSearchContexts={fetchSearchContexts}
                    className={menuClassName}
                    onMenuClose={handleMenuClose}
                />
            </PopoverContent>
        </Popover>
    )
}
