import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Dropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import Shepherd from 'shepherd.js'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { CaseSensitivityProps, PatternTypeProps, SearchContextInputProps } from '..'
import { SubmitSearchParameters } from '../helpers'

import { SearchContextMenu } from './SearchContextMenu'
import { defaultTourOptions } from './tour-options'

export interface SearchContextDropdownProps
    extends Omit<SearchContextInputProps, 'showSearchContext'>,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        VersionContextProps {
    submitSearch: (args: SubmitSearchParameters) => void
    submitSearchOnSearchContextChange?: boolean
    query: string
    history: H.History
    isSearchOnboardingTourVisible: boolean
    className?: string
}

const tourOptions: Shepherd.Tour.TourOptions = {
    ...defaultTourOptions,
    defaultStepOptions: {
        ...defaultTourOptions.defaultStepOptions,
        arrow: true,
        popperOptions: {
            // Removes default behavior of autofocusing steps
            modifiers: [
                {
                    name: 'focusAfterRender',
                    enabled: false,
                },
                { name: 'offset', options: { offset: [2, 4] } },
            ],
        },
    },
}

function getHighlightTourStep(onClose: () => void): HTMLElement {
    const container = document.createElement('div')
    container.className = 'search-context-highlight-tour__step'
    container.innerHTML = `
        <div>
            <strong>New: Search contexts</strong>
        </div>
        <div class="mt-2 mb-2">Search just the code you care about with search contexts.</div>
        <div>
            <a href="https://docs.sourcegraph.com/code_search/explanations/features#search-contexts-experimental" target="_blank">
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
    if (button) {
        button.addEventListener('click', onClose)
    }
    return container
}

const HAS_SEEN_HIGHLIGHT_TOUR_STEP_KEY = 'has-seen-search-contexts-dropdown-highlight-tour-step'

const useSearchContextHighlightTour = (
    showSearchContextHighlightTourStep: boolean,
    isSearchOnboardingTourVisible: boolean
): Shepherd.Tour => {
    const [hasSeenHighlightTourStep, setHasSeenHighlightTourStep] = useLocalStorage(
        HAS_SEEN_HIGHLIGHT_TOUR_STEP_KEY,
        false
    )

    const tour = useMemo(() => new Shepherd.Tour(tourOptions), [])
    useEffect(() => {
        tour.addSteps([
            {
                id: 'search-contexts-start-tour',
                text: getHighlightTourStep(() => tour.cancel()),
                classes: 'web-content shadow-lg py-4 px-3 search-context-highlight-tour',
                attachTo: {
                    element: '.search-context-dropdown__button',
                    on: 'bottom',
                },
                popperOptions: {
                    modifiers: [{ name: 'offset', options: { offset: [140, 16] } }],
                },
            },
        ])
    }, [tour])

    useEffect(() => {
        if (
            !tour.isActive() &&
            showSearchContextHighlightTourStep &&
            !hasSeenHighlightTourStep &&
            !isSearchOnboardingTourVisible
        ) {
            tour.start()
        }
    }, [showSearchContextHighlightTourStep, isSearchOnboardingTourVisible, hasSeenHighlightTourStep, tour])

    useEffect(() => {
        const onCanceled = (): void => {
            setHasSeenHighlightTourStep(true)
        }
        tour.on('cancel', onCanceled)
        return () => {
            tour.off('cancel', onCanceled)
        }
    }, [tour, setHasSeenHighlightTourStep])

    useEffect(() => () => tour.cancel(), [tour])

    return tour
}

export const SearchContextDropdown: React.FunctionComponent<SearchContextDropdownProps> = props => {
    const {
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
        showSearchContextHighlightTourStep = false,
        submitSearchOnSearchContextChange = true,
        isSearchOnboardingTourVisible,
        className,
    } = props

    const tour = useSearchContextHighlightTour(showSearchContextHighlightTourStep, isSearchOnboardingTourVisible)

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
                <SearchContextMenu
                    {...props}
                    selectSearchContextSpec={selectSearchContextSpec}
                    fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                    fetchSearchContexts={fetchSearchContexts}
                    closeMenu={toggleOpen}
                />
            </DropdownMenu>
        </Dropdown>
    )
}
