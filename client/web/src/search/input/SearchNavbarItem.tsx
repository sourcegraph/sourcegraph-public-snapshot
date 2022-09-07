import React, { useCallback, useEffect, useMemo, useRef } from 'react'

import * as H from 'history'
import { of } from 'rxjs'
import { startWith } from 'rxjs/operators'
import shallow from 'zustand/shallow'

import { Form } from '@sourcegraph/branded/src/components/Form'
import {
    InitialParametersSource,
    isSearchContextSpecAvailable,
    SearchContextInputProps,
    SubmitSearchParameters,
} from '@sourcegraph/search'
import { SearchBox } from '@sourcegraph/search-ui'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { getGlobalSearchContextFilter } from '@sourcegraph/shared/src/search/query/query'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { useCoreWorkflowImprovementsEnabled } from '@sourcegraph/shared/src/settings/useCoreWorkflowImprovementsEnabled'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/wildcard'

import { parseSearchURL, parseSearchURLQuery } from '..'
import { AuthenticatedUser } from '../../auth'
import { useExperimentalFeatures, useNavbarQueryState, setSearchCaseSensitivity } from '../../stores'
import { NavbarQueryState, setSearchPatternType } from '../../stores/navbarSearchQueryState'

export interface SearchNavbarItemProps
    extends ActivationProps,
        SettingsCascadeProps,
        ThemeProps,
        SearchContextInputProps,
        TelemetryProps,
        PlatformContextProps<'requestGraphQL'> {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    globbing: boolean
    isSearchAutoFocusRequired?: boolean
    isRepositoryRelatedPage?: boolean
    onHandleFuzzyFinder?: React.Dispatch<React.SetStateAction<boolean>>
}

const selectQueryState = ({
    queryState,
    setQueryState,
    submitSearch,
    searchCaseSensitivity,
    searchPatternType,
}: NavbarQueryState): Pick<
    NavbarQueryState,
    'queryState' | 'setQueryState' | 'submitSearch' | 'searchCaseSensitivity' | 'searchPatternType'
> => ({ queryState, setQueryState, submitSearch, searchCaseSensitivity, searchPatternType })

/**
 * The search item in the navbar
 */
export const SearchNavbarItem: React.FunctionComponent<React.PropsWithChildren<SearchNavbarItemProps>> = (
    props: SearchNavbarItemProps
) => {
    const { history, location, searchContextsEnabled, selectedSearchContextSpec, setSelectedSearchContextSpec } = props
    const autoFocus = props.isSearchAutoFocusRequired ?? true
    // This uses the same logic as in Layout.tsx until we have a better solution
    // or remove the search help button
    const isSearchPage = location.pathname === '/search' && Boolean(parseSearchURLQuery(location.search))

    // Features and settings

    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
    const showSearchContextManagement = useExperimentalFeatures(
        features => features.showSearchContextManagement ?? false
    )
    const editorComponent = useExperimentalFeatures(features => features.editor ?? 'codemirror6')
    const applySuggestionsOnEnter = useExperimentalFeatures(
        features => features.applySearchQuerySuggestionOnEnter ?? false
    )
    const [enableCoreWorkflowImprovements] = useCoreWorkflowImprovementsEnabled()

    // (initial) search state from URL

    const updatingFromURL = useRef(false)
    const {
        query: queryFromURL = '',
        patternType: patternTypeFromURL,
        caseSensitive: caseSensitiveFromURL,
    } = useMemo(() => {
        updatingFromURL.current = true
        return parseSearchURL(location.search)
    }, [location.search])

    const globalSearchContextSpec = useMemo(() => getGlobalSearchContextFilter(queryFromURL), [queryFromURL])
    const isSearchContextAvailable =
        useObservable(
            useMemo(
                () =>
                    globalSearchContextSpec && searchContextsEnabled
                        ? // While we wait for the result of the `isSearchContextSpecAvailable` call, we assume the context is available
                          // to prevent flashing and moving content in the query bar. This optimizes for the most common use case where
                          // user selects a search context from the dropdown.
                          // See https://github.com/sourcegraph/sourcegraph/issues/19918 for more info.
                          isSearchContextSpecAvailable({
                              spec: globalSearchContextSpec.spec,
                              platformContext: props.platformContext,
                          }).pipe(startWith(true))
                        : of(false),
                [globalSearchContextSpec, searchContextsEnabled, props.platformContext]
            )
        ) ?? true // ?? true is necessary because useObservable initially returns undefined

    useEffect(() => {
        // Only override filters from URL if there is a search query
        if (queryFromURL) {
            if (globalSearchContextSpec?.spec && globalSearchContextSpec.spec !== selectedSearchContextSpec) {
                setSelectedSearchContextSpec(globalSearchContextSpec.spec)
            }
        }
    }, [selectedSearchContextSpec, queryFromURL, setSelectedSearchContextSpec, globalSearchContextSpec?.spec])

    const queryFromURLClean = useMemo(
        () =>
            // If a global search context spec is available to the user, we omit it from the
            // query and move it to the search contexts dropdown
            globalSearchContextSpec && isSearchContextAvailable && showSearchContext
                ? omitFilter(queryFromURL, globalSearchContextSpec.filter)
                : queryFromURL,
        [queryFromURL, globalSearchContextSpec, isSearchContextAvailable, showSearchContext]
    )

    // Update internal search state from URL
    useEffect(() => {
        useNavbarQueryState.setState(({ searchPatternType, searchCaseSensitivity }) => ({
            searchPatternType: queryFromURLClean !== '' && patternTypeFromURL ? patternTypeFromURL : searchPatternType,
            searchCaseSensitivity: queryFromURLClean !== '' ? caseSensitiveFromURL : searchCaseSensitivity,
            queryState: { query: queryFromURLClean },
            parametersSource: InitialParametersSource.URL,
            searchQueryFromURL: queryFromURL,
        }))
    }, [queryFromURL, queryFromURLClean, patternTypeFromURL, caseSensitiveFromURL])

    // Internal search state

    const { setQueryState, submitSearch, ...inputProps } = useNavbarQueryState(selectQueryState, shallow)
    let { queryState, searchCaseSensitivity, searchPatternType } = inputProps

    // Although we sync the information from the URL to the internal state, the
    // syncing happens *after* the component rendered. To avoid flashes of
    // content in the query input we use the information directly from the URL
    // when we know that the URL changed.
    if (updatingFromURL.current) {
        if (queryFromURLClean !== '') {
            searchCaseSensitivity = caseSensitiveFromURL
            if (patternTypeFromURL !== undefined) {
                searchPatternType = patternTypeFromURL
            }
        }
        queryState = { query: queryFromURLClean }
        updatingFromURL.current = false
    }

    const submitSearchOnChange = useCallback(
        (parameters: Partial<SubmitSearchParameters> = {}) => {
            submitSearch({
                history,
                source: 'nav',
                activation: props.activation,
                selectedSearchContextSpec: props.selectedSearchContextSpec,
                ...parameters,
            })
        },
        [submitSearch, history, props.activation, props.selectedSearchContextSpec]
    )

    const onSubmit = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            submitSearchOnChange()
        },
        [submitSearchOnChange]
    )

    return (
        <Form
            className="search--navbar-item d-flex align-items-flex-start flex-grow-1 flex-shrink-past-contents"
            onSubmit={onSubmit}
        >
            <SearchBox
                {...props}
                editorComponent={editorComponent}
                applySuggestionsOnEnter={enableCoreWorkflowImprovements || applySuggestionsOnEnter}
                showSearchContext={showSearchContext}
                showSearchContextManagement={showSearchContextManagement}
                caseSensitive={searchCaseSensitivity}
                setCaseSensitivity={setSearchCaseSensitivity}
                patternType={searchPatternType}
                setPatternType={setSearchPatternType}
                queryState={queryState}
                onChange={setQueryState}
                onSubmit={onSubmit}
                submitSearchOnToggle={submitSearchOnChange}
                submitSearchOnSearchContextChange={submitSearchOnChange}
                autoFocus={autoFocus}
                hideHelpButton={isSearchPage}
                onHandleFuzzyFinder={props.onHandleFuzzyFinder}
                isExternalServicesUserModeAll={window.context.externalServicesUserMode === 'all'}
                structuralSearchDisabled={window.context?.experimentalFeatures?.structuralSearch === 'disabled'}
            />
        </Form>
    )
}
