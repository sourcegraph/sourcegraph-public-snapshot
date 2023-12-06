import React, { useCallback, useRef, useEffect } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'
import shallow from 'zustand/shallow'

import { SearchBox, Toggles } from '@sourcegraph/branded'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SearchContextInputProps, SubmitSearchParameters } from '@sourcegraph/shared/src/search'
import { type SettingsCascadeProps, useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Form } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { useNavbarQueryState, setSearchCaseSensitivity } from '../../stores'
import { type NavbarQueryState, setSearchMode, setSearchPatternType } from '../../stores/navbarSearchQueryState'
import { useExperimentalQueryInput } from '../useExperimentalSearchInput'

import { LazyExperimentalSearchInput } from './LazyExperimentalSearchInput'
import { useRecentSearches } from './useRecentSearches'

interface Props
    extends SettingsCascadeProps,
        SearchContextInputProps,
        TelemetryProps,
        TelemetryV2Props,
        PlatformContextProps<'requestGraphQL'> {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    isSearchAutoFocusRequired?: boolean
    isRepositoryRelatedPage?: boolean
    isLightTheme: boolean
}

const selectQueryState = ({
    queryState,
    setQueryState,
    submitSearch,
    searchCaseSensitivity,
    searchPatternType,
    searchMode,
}: NavbarQueryState): Pick<
    NavbarQueryState,
    'queryState' | 'setQueryState' | 'submitSearch' | 'searchCaseSensitivity' | 'searchPatternType' | 'searchMode'
> => ({ queryState, setQueryState, submitSearch, searchCaseSensitivity, searchPatternType, searchMode })

/**
 * The search item in the navbar
 */
export const SearchNavbarItem: React.FunctionComponent<React.PropsWithChildren<Props>> = (props: Props) => {
    const navigate = useNavigate()
    const location = useLocation()

    const { queryState, setQueryState, submitSearch, searchCaseSensitivity, searchPatternType, searchMode } =
        useNavbarQueryState(selectQueryState, shallow)

    const [experimentalQueryInput] = useExperimentalQueryInput()
    const applySuggestionsOnEnter =
        useExperimentalFeatures(features => features.applySearchQuerySuggestionOnEnter) ?? true

    const { recentSearches } = useRecentSearches()

    const submitSearchOnChange = useCallback(
        (parameters: Partial<SubmitSearchParameters> = {}) => {
            submitSearch({
                historyOrNavigate: navigate,
                location,
                source: 'nav',
                selectedSearchContextSpec: props.selectedSearchContextSpec,
                ...parameters,
            })
        },
        [submitSearch, navigate, location, props.selectedSearchContextSpec]
    )
    const submitSearchOnChangeRef = useRef(submitSearchOnChange)
    useEffect(() => {
        submitSearchOnChangeRef.current = submitSearchOnChange
    }, [submitSearchOnChange])

    const onSubmit = useCallback((event?: React.FormEvent): void => {
        event?.preventDefault()
        submitSearchOnChangeRef.current()
    }, [])

    // TODO (#48103): Remove/simplify when new search input is released
    if (experimentalQueryInput) {
        return (
            <Form
                className="search--navbar-item d-flex align-items-flex-start flex-grow-1 flex-shrink-past-contents"
                onSubmit={onSubmit}
            >
                <LazyExperimentalSearchInput
                    visualMode="compact"
                    telemetryService={props.telemetryService}
                    telemetryRecorder={props.telemetryRecorder}
                    patternType={searchPatternType}
                    interpretComments={false}
                    queryState={queryState}
                    onChange={setQueryState}
                    onSubmit={onSubmit}
                    isLightTheme={props.isLightTheme}
                    platformContext={props.platformContext}
                    authenticatedUser={props.authenticatedUser}
                    fetchSearchContexts={props.fetchSearchContexts}
                    getUserSearchContextNamespaces={props.getUserSearchContextNamespaces}
                    isSourcegraphDotCom={props.isSourcegraphDotCom}
                    submitSearch={submitSearchOnChange}
                    selectedSearchContextSpec={props.selectedSearchContextSpec}
                    className="flex-grow-1"
                >
                    <Toggles
                        patternType={searchPatternType}
                        caseSensitive={searchCaseSensitivity}
                        setPatternType={setSearchPatternType}
                        setCaseSensitivity={setSearchCaseSensitivity}
                        searchMode={searchMode}
                        setSearchMode={setSearchMode}
                        settingsCascade={props.settingsCascade}
                        navbarSearchQuery={queryState.query}
                        submitSearch={submitSearchOnChange}
                        structuralSearchDisabled={window.context?.experimentalFeatures?.structuralSearch === 'disabled'}
                    />
                </LazyExperimentalSearchInput>
            </Form>
        )
    }

    return (
        <Form
            className="search--navbar-item d-flex align-items-flex-start flex-grow-1 flex-shrink-past-contents"
            onSubmit={onSubmit}
        >
            <SearchBox
                {...props}
                autoFocus={false}
                applySuggestionsOnEnter={applySuggestionsOnEnter}
                showSearchContext={props.searchContextsEnabled}
                showSearchContextManagement={true}
                caseSensitive={searchCaseSensitivity}
                setCaseSensitivity={setSearchCaseSensitivity}
                patternType={searchPatternType}
                setPatternType={setSearchPatternType}
                searchMode={searchMode}
                setSearchMode={setSearchMode}
                queryState={queryState}
                onChange={setQueryState}
                onSubmit={onSubmit}
                submitSearchOnToggle={submitSearchOnChange}
                submitSearchOnSearchContextChange={submitSearchOnChange}
                isExternalServicesUserModeAll={window.context.externalServicesUserMode === 'all'}
                structuralSearchDisabled={window.context?.experimentalFeatures?.structuralSearch === 'disabled'}
                hideHelpButton={false}
                showSearchHistory={true}
                recentSearches={recentSearches}
            />
        </Form>
    )
}
