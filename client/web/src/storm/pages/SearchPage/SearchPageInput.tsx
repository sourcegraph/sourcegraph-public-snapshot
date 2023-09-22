import React, { type FC, useCallback, useEffect, useRef } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'
import type { NavbarQueryState } from 'src/stores/navbarSearchQueryState'
import shallow from 'zustand/shallow'

import { Toggles } from '@sourcegraph/branded'
import { TraceSpanProvider } from '@sourcegraph/observability-client'
import {
    type CaseSensitivityProps,
    type SearchPatternTypeProps,
    type SubmitSearchParameters,
    canSubmitSearch,
    type QueryState,
    type SearchModeProps,
    getUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/search'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Form } from '@sourcegraph/wildcard'

import { Notices } from '../../../global/Notices'
import { useLegacyContext_onlyInStormRoutes } from '../../../LegacyRouteContext'
import { submitSearch } from '../../../search/helpers'
import { LazyExperimentalSearchInput } from '../../../search/input/LazyExperimentalSearchInput'
import { useNavbarQueryState, setSearchCaseSensitivity, setSearchPatternType, setSearchMode } from '../../../stores'

import { SimpleSearch } from './SimpleSearch'

import styles from './SearchPageInput.module.scss'

const queryStateSelector = (
    state: NavbarQueryState
): Pick<CaseSensitivityProps, 'caseSensitive'> & SearchPatternTypeProps & Pick<SearchModeProps, 'searchMode'> => ({
    caseSensitive: state.searchCaseSensitivity,
    patternType: state.searchPatternType,
    searchMode: state.searchMode,
})

interface SearchPageInputProps {
    queryState: QueryState
    setQueryState: (newState: QueryState) => void
    hardCodedSearchContextSpec?: string
    simpleSearch: boolean
}

export const SearchPageInput: FC<SearchPageInputProps> = props => {
    const { queryState, setQueryState, hardCodedSearchContextSpec, simpleSearch } = props

    const {
        authenticatedUser,
        isSourcegraphDotCom,
        telemetryService,
        platformContext,
        settingsCascade,
        selectedSearchContextSpec: dynamicSearchContextSpec,
        fetchSearchContexts,
    } = useLegacyContext_onlyInStormRoutes()

    const selectedSearchContextSpec = hardCodedSearchContextSpec || dynamicSearchContextSpec

    const location = useLocation()
    const navigate = useNavigate()

    const isLightTheme = useIsLightTheme()
    const { caseSensitive, patternType, searchMode } = useNavbarQueryState(queryStateSelector, shallow)

    const submitSearchOnChange = useCallback(
        (parameters: Partial<SubmitSearchParameters> = {}) => {
            const query = parameters.query ?? queryState.query

            if (canSubmitSearch(query, selectedSearchContextSpec)) {
                submitSearch({
                    source: 'home',
                    query,
                    historyOrNavigate: navigate,
                    location,
                    patternType,
                    caseSensitive,
                    searchMode,
                    ...parameters,
                })
            }
        },
        [queryState.query, selectedSearchContextSpec, navigate, location, patternType, caseSensitive, searchMode]
    )
    const submitSearchOnChangeRef = useRef(submitSearchOnChange)
    useEffect(() => {
        submitSearchOnChangeRef.current = submitSearchOnChange
    }, [submitSearchOnChange])

    const onSubmit = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            submitSearchOnChangeRef.current()
        },
        [submitSearchOnChangeRef]
    )

    const onSimpleSearchUpdate = useCallback(
        (val: string) => {
            setQueryState({ query: val })
        },
        [setQueryState]
    )

    const input = (
        <LazyExperimentalSearchInput
            telemetryService={telemetryService}
            patternType={patternType}
            interpretComments={false}
            queryState={queryState}
            onChange={setQueryState}
            onSubmit={onSubmit}
            isLightTheme={isLightTheme}
            platformContext={platformContext}
            authenticatedUser={authenticatedUser}
            fetchSearchContexts={fetchSearchContexts}
            getUserSearchContextNamespaces={getUserSearchContextNamespaces}
            isSourcegraphDotCom={isSourcegraphDotCom}
            submitSearch={submitSearchOnChange}
            selectedSearchContextSpec={selectedSearchContextSpec}
            className="flex-grow-1"
        >
            <Toggles
                patternType={patternType}
                caseSensitive={caseSensitive}
                setPatternType={setSearchPatternType}
                setCaseSensitivity={setSearchCaseSensitivity}
                searchMode={searchMode}
                setSearchMode={setSearchMode}
                settingsCascade={settingsCascade}
                navbarSearchQuery={queryState.query}
                showCopyQueryButton={false}
                showSmartSearchButton={false}
                structuralSearchDisabled={window.context?.experimentalFeatures?.structuralSearch === 'disabled'}
            />
        </LazyExperimentalSearchInput>
    )

    return (
        <div>
            <div className="d-flex flex-row flex-shrink-past-contents">
                <Form className="flex-grow-1 flex-shrink-past-contents" onSubmit={onSubmit}>
                    <div data-search-page-input-container={true} className={styles.inputContainer}>
                        <TraceSpanProvider name="SearchBox">
                            <div className="d-flex flex-grow-1 w-100">{input}</div>
                        </TraceSpanProvider>
                    </div>
                    <Notices className="my-3 text-center" location="home" />
                </Form>
            </div>
            {simpleSearch && (
                <div>
                    <hr className="mt-4 mb-4" />
                    <SimpleSearch
                        telemetryService={telemetryService}
                        onSubmit={onSubmit}
                        onSimpleSearchUpdate={onSimpleSearchUpdate}
                    />
                </div>
            )}
        </div>
    )
}
