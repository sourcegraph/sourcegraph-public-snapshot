import React, { useCallback } from 'react'

import * as H from 'history'
import shallow from 'zustand/shallow'

import { SearchBox } from '@sourcegraph/branded'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SearchContextInputProps, SubmitSearchParameters } from '@sourcegraph/shared/src/search'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Form } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { useExperimentalFeatures, useNavbarQueryState, setSearchCaseSensitivity } from '../../stores'
import { NavbarQueryState, setSearchMode, setSearchPatternType } from '../../stores/navbarSearchQueryState'

import { useRecentSearches } from './useRecentSearches'

interface Props
    extends SettingsCascadeProps,
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
    const { queryState, setQueryState, submitSearch, searchCaseSensitivity, searchPatternType, searchMode } =
        useNavbarQueryState(selectQueryState, shallow)

    const applySuggestionsOnEnter =
        useExperimentalFeatures(features => features.applySearchQuerySuggestionOnEnter) ?? true

    const { recentSearches } = useRecentSearches()

    const submitSearchOnChange = useCallback(
        (parameters: Partial<SubmitSearchParameters> = {}) => {
            submitSearch({
                history: props.history,
                source: 'nav',
                selectedSearchContextSpec: props.selectedSearchContextSpec,
                ...parameters,
            })
        },
        [submitSearch, props.history, props.selectedSearchContextSpec]
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
