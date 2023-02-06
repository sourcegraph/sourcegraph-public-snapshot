import React, { useCallback, useMemo, useRef } from 'react'

import { useLocation, useNavigate } from 'react-router-dom-v5-compat'
import { NavbarQueryState } from 'src/stores/navbarSearchQueryState'
import shallow from 'zustand/shallow'

import { SearchBox, Toggles } from '@sourcegraph/branded'
// The experimental search input should be shown on the search home page
// eslint-disable-next-line no-restricted-imports
import { LazyCodeMirrorQueryInput, searchHistoryExtension } from '@sourcegraph/branded/src/search-ui/experimental'
import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import {
    SearchContextInputProps,
    CaseSensitivityProps,
    SearchPatternTypeProps,
    SubmitSearchParameters,
    canSubmitSearch,
    QueryState,
    SearchModeProps,
    SearchContextProps,
} from '@sourcegraph/shared/src/search'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Form } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { Notices } from '../../global/Notices'
import {
    useExperimentalFeatures,
    useNavbarQueryState,
    setSearchCaseSensitivity,
    setSearchPatternType,
    setSearchMode,
} from '../../stores'
import { ThemePreferenceProps } from '../../theme'
import { submitSearch } from '../helpers'
import { createSuggestionsSource } from '../input/suggestions'
import { useRecentSearches } from '../input/useRecentSearches'

import styles from './SearchPageInput.module.scss'

interface Props
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        TelemetryProps,
        PlatformContextProps<'settings' | 'sourcegraphURL' | 'requestGraphQL'>,
        Pick<SubmitSearchParameters, 'source'>,
        SearchContextInputProps,
        Pick<SearchContextProps, 'searchContextsEnabled'> {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    /** Whether globbing is enabled for filters. */
    globbing: boolean
    autoFocus?: boolean
    queryState: QueryState
    setQueryState: (newState: QueryState) => void
}

const queryStateSelector = (
    state: NavbarQueryState
): Pick<CaseSensitivityProps, 'caseSensitive'> & SearchPatternTypeProps & Pick<SearchModeProps, 'searchMode'> => ({
    caseSensitive: state.searchCaseSensitivity,
    patternType: state.searchPatternType,
    searchMode: state.searchMode,
})

export const SearchPageInput: React.FunctionComponent<React.PropsWithChildren<Props>> = (props: Props) => {
    const location = useLocation()
    const navigate = useNavigate()

    const { caseSensitive, patternType, searchMode } = useNavbarQueryState(queryStateSelector, shallow)
    const experimentalQueryInput = useExperimentalFeatures(features => features.searchQueryInput === 'experimental')
    const applySuggestionsOnEnter =
        useExperimentalFeatures(features => features.applySearchQuerySuggestionOnEnter) ?? true

    const { recentSearches } = useRecentSearches()
    const recentSearchesRef = useRef(recentSearches)
    recentSearchesRef.current = recentSearches

    const submitSearchOnChange = useCallback(
        (parameters: Partial<SubmitSearchParameters> = {}) => {
            const query = props.queryState.query

            if (canSubmitSearch(query, props.selectedSearchContextSpec)) {
                submitSearch({
                    source: 'home',
                    query,
                    historyOrNavigate: navigate,
                    location,
                    patternType,
                    caseSensitive,
                    searchMode,
                    // In the new query input, context is either omitted (-> global)
                    // or explicitly specified.
                    selectedSearchContextSpec: experimentalQueryInput ? undefined : props.selectedSearchContextSpec,
                    ...parameters,
                })
            }
        },
        [
            props.queryState.query,
            props.selectedSearchContextSpec,
            navigate,
            location,
            patternType,
            caseSensitive,
            searchMode,
            experimentalQueryInput,
        ]
    )
    const submitSearchOnChangeRef = useRef(submitSearchOnChange)
    submitSearchOnChangeRef.current = submitSearchOnChange

    const onSubmit = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            submitSearchOnChangeRef.current()
        },
        [submitSearchOnChangeRef]
    )

    // We want to prevent autofocus by default on devices with touch as their only input method.
    // Touch only devices result in the onscreen keyboard not showing until the input loses focus and
    // gets focused again by the user. The logic is not fool proof, but should rule out majority of cases
    // where a touch enabled device has a physical keyboard by relying on detection of a fine pointer with hover ability.
    const isTouchOnlyDevice =
        !window.matchMedia('(any-pointer:fine)').matches && window.matchMedia('(any-hover:none)').matches

    const suggestionSource = useMemo(
        () =>
            createSuggestionsSource({
                platformContext: props.platformContext,
                authenticatedUser: props.authenticatedUser,
                fetchSearchContexts: props.fetchSearchContexts,
                getUserSearchContextNamespaces: props.getUserSearchContextNamespaces,
                isSourcegraphDotCom: props.isSourcegraphDotCom,
            }),
        [
            props.platformContext,
            props.authenticatedUser,
            props.fetchSearchContexts,
            props.getUserSearchContextNamespaces,
            props.isSourcegraphDotCom,
        ]
    )

    const experimentalExtensions = useMemo(
        () =>
            experimentalQueryInput
                ? [
                      searchHistoryExtension({
                          mode: {
                              name: 'History',
                              placeholder: 'Filter history',
                          },
                          source: () => recentSearchesRef.current ?? [],
                          submitQuery: query => submitSearchOnChangeRef.current({ query }),
                      }),
                  ]
                : [],
        [experimentalQueryInput, recentSearchesRef, submitSearchOnChangeRef]
    )

    const input = experimentalQueryInput ? (
        <LazyCodeMirrorQueryInput
            patternType={patternType}
            interpretComments={false}
            queryState={props.queryState}
            onChange={props.setQueryState}
            onSubmit={onSubmit}
            isLightTheme={props.isLightTheme}
            placeholder="Search for code or files..."
            suggestionSource={suggestionSource}
            extensions={experimentalExtensions}
        >
            <Toggles
                patternType={patternType}
                caseSensitive={caseSensitive}
                setPatternType={setSearchPatternType}
                setCaseSensitivity={setSearchCaseSensitivity}
                searchMode={searchMode}
                setSearchMode={setSearchMode}
                settingsCascade={props.settingsCascade}
                navbarSearchQuery={props.queryState.query}
            />
        </LazyCodeMirrorQueryInput>
    ) : (
        <SearchBox
            {...props}
            showSearchContext={props.searchContextsEnabled}
            showSearchContextManagement={true}
            caseSensitive={caseSensitive}
            patternType={patternType}
            setPatternType={setSearchPatternType}
            setCaseSensitivity={setSearchCaseSensitivity}
            searchMode={searchMode}
            setSearchMode={setSearchMode}
            queryState={props.queryState}
            onChange={props.setQueryState}
            onSubmit={onSubmit}
            autoFocus={!isTouchOnlyDevice && props.autoFocus !== false}
            isExternalServicesUserModeAll={window.context.externalServicesUserMode === 'all'}
            structuralSearchDisabled={window.context?.experimentalFeatures?.structuralSearch === 'disabled'}
            applySuggestionsOnEnter={applySuggestionsOnEnter}
            showSearchHistory={true}
            recentSearches={recentSearches}
        />
    )
    return (
        <div className="d-flex flex-row flex-shrink-past-contents">
            <Form className="flex-grow-1 flex-shrink-past-contents" onSubmit={onSubmit}>
                <div data-search-page-input-container={true} className={styles.inputContainer}>
                    <TraceSpanProvider name="SearchBox">
                        <div className="d-flex flex-grow-1 w-100">{input}</div>
                    </TraceSpanProvider>
                </div>
                <Notices className="my-3 text-center" location="home" settingsCascade={props.settingsCascade} />
            </Form>
        </div>
    )
}
