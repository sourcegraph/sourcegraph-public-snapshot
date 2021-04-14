import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Dropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import Shepherd from 'shepherd.js'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { filterExists } from '@sourcegraph/shared/src/search/query/validate'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { CaseSensitivityProps, PatternTypeProps, SearchContextProps } from '..'
import { SubmitSearchParameters } from '../helpers'

import { SearchContextMenu } from './SearchContextMenu'
import { defaultTourOptions } from './tour-options'

export interface SearchContextDropdownProps
    extends Omit<SearchContextProps, 'showSearchContext'>,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        VersionContextProps {
    submitSearch: (args: SubmitSearchParameters) => void
    submitSearchOnSearchContextChange?: boolean
    query: string
    history: H.History
}

const tourOptions: Shepherd.Tour.TourOptions = {
    ...defaultTourOptions,
    defaultStepOptions: {
        ...defaultTourOptions.defaultStepOptions,
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
    container.className = 'search-context-dropdown__highlight-tour-step'
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
    } = props

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
                attachTo: {
                    element: '.search-context-dropdown__button',
                    on: 'bottom',
                },
            },
        ])
    }, [tour])

    useEffect(() => {
        if (showSearchContextHighlightTourStep && !hasSeenHighlightTourStep) {
            tour.start()
        }
    }, [showSearchContextHighlightTourStep, hasSeenHighlightTourStep, tour])

    useEffect(() => {
        const onCanceled = (): void => {
            setHasSeenHighlightTourStep(true)
        }
        tour.on('cancel', onCanceled)
        return () => {
            tour.off('cancel', onCanceled)
        }
    }, [tour, setHasSeenHighlightTourStep])

    useEffect(
        () => () => {
            if (tour.isActive()) {
                tour.cancel()
            }
        },
        [tour]
    )

    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => {
        setIsOpen(value => !value)
        tour.cancel()
    }, [tour])

    const [isDisabled, setIsDisabled] = useState(false)

    // Disable the dropdown if the query contains a context filter
    useEffect(() => setIsDisabled(filterExists(query, FilterType.context)), [query])

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
        <>
            <Dropdown
                isOpen={isOpen}
                toggle={toggleOpen}
                a11y={false} /* Override default keyboard events in reactstrap */
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
                    data-tooltip={isDisabled ? 'Overridden by query' : ''}
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
                <DropdownMenu>
                    <SearchContextMenu
                        {...props}
                        selectSearchContextSpec={selectSearchContextSpec}
                        fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                        fetchSearchContexts={fetchSearchContexts}
                        closeMenu={toggleOpen}
                    />
                </DropdownMenu>
            </Dropdown>
            <div className="search-context-dropdown__separator" />
        </>
    )
}
