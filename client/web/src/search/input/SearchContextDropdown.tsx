import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Dropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CaseSensitivityProps, PatternTypeProps, SearchContextInputProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { useTemporarySetting } from '../../settings/temporary/useTemporarySetting'
import { SubmitSearchParameters } from '../helpers'

import { SearchContextCtaPrompt } from './SearchContextCtaPrompt'
import { SearchContextMenu } from './SearchContextMenu'

export interface SearchContextDropdownProps
    extends Omit<SearchContextInputProps, 'showSearchContext'>,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        TelemetryProps {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
    submitSearch: (args: SubmitSearchParameters) => void
    submitSearchOnSearchContextChange?: boolean
    query: string
    history: H.History
    className?: string
}

export const SearchContextDropdown: React.FunctionComponent<SearchContextDropdownProps> = props => {
    const {
        isSourcegraphDotCom,
        authenticatedUser,
        hasUserAddedRepositories,
        hasUserAddedExternalServices,
        history,
        patternType,
        caseSensitive,
        query,
        selectedSearchContextSpec,
        setSelectedSearchContextSpec,
        submitSearch,
        fetchAutoDefinedSearchContexts,
        fetchSearchContexts,
        submitSearchOnSearchContextChange = true,
        className,
        telemetryService,
    } = props

    const [hasUsedNonGlobalContext] = useTemporarySetting('search.usedNonGlobalContext')

    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => {
        telemetryService.log('SearchContextDropdownToggled')
        setIsOpen(value => !value)
    }, [telemetryService])

    const isContextFilterInQuery = useMemo(() => filterExists(query, FilterType.context), [query])

    const disabledTooltipText = isContextFilterInQuery ? 'Overridden by query' : ''

    const submitOnToggle = useCallback(
        (selectedSearchContextSpec: string): void => {
            submitSearch({
                history,
                query,
                source: 'filter',
                patternType,
                caseSensitive,
                selectedSearchContextSpec,
            })
        },
        [submitSearch, caseSensitive, history, query, patternType]
    )

    const selectSearchContextSpec = useCallback(
        (spec: string): void => {
            if (submitSearchOnSearchContextChange) {
                submitOnToggle(spec)
            } else {
                setSelectedSearchContextSpec(spec)
            }
        },
        [submitSearchOnSearchContextChange, submitOnToggle, setSelectedSearchContextSpec]
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

    return (
        <Dropdown
            isOpen={isOpen}
            toggle={toggleOpen}
            a11y={false} /* Override default keyboard events in reactstrap */
            className={classNames('search-context-dropdown ', className)}
        >
            <DropdownToggle
                className={classNames(
                    'search-context-dropdown__button',
                    'dropdown-toggle',
                    'test-search-context-dropdown',
                    {
                        'search-context-dropdown__button--open': isOpen,
                    }
                )}
                color="link"
                disabled={isContextFilterInQuery}
                data-tooltip={disabledTooltipText}
            >
                <code className="search-context-dropdown__button-content test-selected-search-context-spec">
                    <span className="search-filter-keyword">context:</span>
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
            <DropdownMenu positionFixed={true} className="search-context-dropdown__menu">
                {isSourcegraphDotCom && !hasUserAddedRepositories && !hasUsedNonGlobalContext ? (
                    <SearchContextCtaPrompt
                        telemetryService={telemetryService}
                        authenticatedUser={authenticatedUser}
                        hasUserAddedExternalServices={hasUserAddedExternalServices}
                    />
                ) : (
                    <SearchContextMenu
                        {...props}
                        selectSearchContextSpec={selectSearchContextSpec}
                        fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                        fetchSearchContexts={fetchSearchContexts}
                        closeMenu={toggleOpen}
                    />
                )}
            </DropdownMenu>
        </Dropdown>
    )
}
