import React, { useCallback, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
// eslint-disable-next-line no-restricted-imports
import { Dropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { SearchContextInputProps, SubmitSearchProps } from '@sourcegraph/search'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { SearchContextCtaPrompt } from './SearchContextCtaPrompt'
import { SearchContextMenu } from './SearchContextMenu'

import styles from './SearchContextDropdown.module.scss'

export interface SearchContextDropdownProps
    extends SearchContextInputProps,
        TelemetryProps,
        Partial<Pick<SubmitSearchProps, 'submitSearch'>>,
        PlatformContextProps<'requestGraphQL'> {
    isSourcegraphDotCom: boolean
    showSearchContextManagement: boolean
    authenticatedUser: AuthenticatedUser | null
    query: string
    className?: string
    onEscapeMenuClose?: () => void
    isExternalServicesUserModeAll?: boolean
}

export const SearchContextDropdown: React.FunctionComponent<
    React.PropsWithChildren<SearchContextDropdownProps>
> = props => {
    const {
        isSourcegraphDotCom,
        authenticatedUser,
        hasUserAddedRepositories,
        hasUserAddedExternalServices,
        query,
        selectedSearchContextSpec,
        setSelectedSearchContextSpec,
        submitSearch,
        fetchAutoDefinedSearchContexts,
        fetchSearchContexts,
        className,
        telemetryService,
        onEscapeMenuClose,
        showSearchContextManagement,
        isExternalServicesUserModeAll,
    } = props

    const [contextCtaDismissed, setContextCtaDismissed] = useTemporarySetting('search.contexts.ctaDismissed', false)

    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => {
        telemetryService.log('SearchContextDropdownToggled')
        setIsOpen(value => !value)
    }, [telemetryService])

    const isContextFilterInQuery = useMemo(() => filterExists(query, FilterType.context), [query])

    const disabledTooltipText = isContextFilterInQuery ? 'Overridden by query' : ''

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

    useEffect(() => {
        if (isOpen && authenticatedUser) {
            // Log search context dropdown view event whenever dropdown is opened, if user is authenticated
            telemetryService.log('SearchContextsDropdownViewed')
        }

        if (isOpen && !authenticatedUser) {
            // Log CTA view event whenver dropdown is opened, if user is not authenticated
            telemetryService.log('SearchResultContextsCTAShown')
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [isOpen])

    const onCtaDismissed = (): void => {
        setContextCtaDismissed(true)
    }
    const isUserAnOrgMember = authenticatedUser && authenticatedUser.organizations.nodes.length !== 0

    const onCloseMenu = useCallback(
        (isEscapeKey?: boolean) => {
            if (isEscapeKey) {
                onEscapeMenuClose?.()
            }
            toggleOpen()
        },
        [toggleOpen, onEscapeMenuClose]
    )

    return (
        <Dropdown
            isOpen={isOpen}
            data-testid="dropdown"
            toggle={toggleOpen}
            a11y={false} /* Override default keyboard events in reactstrap */
            className={className}
        >
            <DropdownToggle
                className={classNames(
                    styles.button,
                    'dropdown-toggle',
                    'test-search-context-dropdown',
                    isOpen && styles.buttonOpen
                )}
                data-testid="dropdown-toggle"
                color="link"
                disabled={isContextFilterInQuery}
                data-tooltip={disabledTooltipText}
            >
                <code className={classNames('test-selected-search-context-spec', styles.buttonContent)}>
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
                </code>
            </DropdownToggle>
            {/*
               a11y-ignore
               Rule: "aria-required-children" (Certain ARIA roles must contain particular children)
               GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/34348
             */}
            <DropdownMenu positionFixed={true} className={classNames('a11y-ignore', styles.menu)}>
                <SearchContextMenu
                    {...props}
                    showSearchContextManagement={showSearchContextManagement}
                    selectSearchContextSpec={selectSearchContextSpec}
                    fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                    fetchSearchContexts={fetchSearchContexts}
                    closeMenu={onCloseMenu}
                />
                {isSourcegraphDotCom && !isUserAnOrgMember && !hasUserAddedRepositories && !contextCtaDismissed && (
                    <SearchContextCtaPrompt
                        telemetryService={telemetryService}
                        authenticatedUser={authenticatedUser}
                        hasUserAddedExternalServices={hasUserAddedExternalServices}
                        onDismiss={onCtaDismissed}
                        isExternalServicesUserModeAll={isExternalServicesUserModeAll}
                    />
                )}
            </DropdownMenu>
        </Dropdown>
    )
}
