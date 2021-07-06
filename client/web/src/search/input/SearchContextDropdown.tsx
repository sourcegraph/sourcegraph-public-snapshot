import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useMemo, useState } from 'react'
import { Dropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CaseSensitivityProps, PatternTypeProps, SearchContextInputProps } from '..'
import { AuthenticatedUser } from '../../auth'
import styles from '../FeatureTour.module.scss'
import { SubmitSearchParameters } from '../helpers'
import { getTourOptions, HAS_SEEN_SEARCH_CONTEXTS_FEATURE_TOUR_KEY, useFeatureTour } from '../useFeatureTour'

import { SearchContextCtaPrompt } from './SearchContextCtaPrompt'
import { SearchContextMenu } from './SearchContextMenu'
import { defaultPopperModifiers } from './tour-options'

export interface SearchContextDropdownProps
    extends Omit<SearchContextInputProps, 'showSearchContext'>,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        VersionContextProps,
        TelemetryProps {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
    submitSearch: (args: SubmitSearchParameters) => void
    submitSearchOnSearchContextChange?: boolean
    query: string
    history: H.History
    isSearchOnboardingTourVisible: boolean
    className?: string
}

function getFeatureTourElement(onClose: () => void): HTMLElement {
    const container = document.createElement('div')
    container.className = styles.featureTourStep
    container.innerHTML = `
        <div>
            <strong>New: Search contexts</strong>
        </div>
        <div class="mt-2 mb-2">Search just the code you care about with search contexts.</div>
        <div>
            <a href="https://docs.sourcegraph.com/code_search/explanations/features#search-contexts" target="_blank">
                Learn more
            </a>
        </div>
        <div class="d-flex justify-content-end">
            <button type="button" class="btn btn-sm">
                Close
            </button>
        </div>
    `
    const button = container.querySelector('button')
    button?.addEventListener('click', onClose)
    return container
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
        versionContext,
        query,
        selectedSearchContextSpec,
        setSelectedSearchContextSpec,
        submitSearch,
        fetchAutoDefinedSearchContexts,
        fetchSearchContexts,
        showSearchContextFeatureTour = false,
        submitSearchOnSearchContextChange = true,
        isSearchOnboardingTourVisible,
        className,
    } = props

    const tour = useFeatureTour(
        'search-contexts-start-tour',
        !!authenticatedUser && showSearchContextFeatureTour && !isSearchOnboardingTourVisible,
        getFeatureTourElement,
        HAS_SEEN_SEARCH_CONTEXTS_FEATURE_TOUR_KEY,
        getTourOptions({
            attachTo: {
                element: '.search-context-dropdown__button',
                on: 'bottom',
            },
            popperOptions: {
                modifiers: [...defaultPopperModifiers, { name: 'offset', options: { offset: [140, 16] } }],
            },
        })
    )

    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => {
        setIsOpen(value => !value)
        tour.cancel()
    }, [tour])

    const isContextFilterInQuery = useMemo(() => filterExists(query, FilterType.context), [query])

    // Disable the dropdown if the query contains a context filter or if a version context is active
    const isDisabled = isContextFilterInQuery || !!versionContext
    const disabledTooltipText = isContextFilterInQuery
        ? 'Overridden by query'
        : versionContext
        ? 'Overriden by version context'
        : ''

    const submitOnToggle = useCallback(
        (selectedSearchContextSpec: string): void => {
            submitSearch({
                history,
                query,
                source: 'filter',
                patternType,
                caseSensitive,
                selectedSearchContextSpec,
                versionContext,
            })
        },
        [submitSearch, caseSensitive, history, query, patternType, versionContext]
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
                disabled={isDisabled}
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
                {isSourcegraphDotCom && !hasUserAddedRepositories ? (
                    <SearchContextCtaPrompt
                        telemetryService={props.telemetryService}
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
