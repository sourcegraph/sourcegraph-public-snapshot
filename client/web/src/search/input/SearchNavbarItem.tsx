import React, { useCallback } from 'react'

import * as H from 'history'
import shallow from 'zustand/shallow'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { SearchContextInputProps, SubmitSearchParameters } from '@sourcegraph/search'
import { SearchBox } from '@sourcegraph/search-ui'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { parseSearchURLQuery } from '..'
import { AuthenticatedUser } from '../../auth'
import { useExperimentalFeatures, useNavbarQueryState, setSearchCaseSensitivity } from '../../stores'
import { NavbarQueryState, setSearchPatternType } from '../../stores/navbarSearchQueryState'

import { useRecentSearches } from './useRecentSearches'

interface Props
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
export const SearchNavbarItem: React.FunctionComponent<React.PropsWithChildren<Props>> = (props: Props) => {
    const autoFocus = props.isSearchAutoFocusRequired ?? true
    // This uses the same logic as in Layout.tsx until we have a better solution
    // or remove the search help button
    const isSearchPage = props.location.pathname === '/search' && Boolean(parseSearchURLQuery(props.location.search))
    const { queryState, setQueryState, submitSearch, searchCaseSensitivity, searchPatternType } = useNavbarQueryState(
        selectQueryState,
        shallow
    )
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
    const showSearchContextManagement = useExperimentalFeatures(
        features => features.showSearchContextManagement ?? false
    )
    const editorComponent = useExperimentalFeatures(features => features.editor ?? 'codemirror6')
    const applySuggestionsOnEnter =
        useExperimentalFeatures(features => features.applySearchQuerySuggestionOnEnter) ?? true

    const { addRecentSearch } = useRecentSearches()

    const submitSearchOnChange = useCallback(
        (parameters: Partial<SubmitSearchParameters> = {}) => {
            submitSearch({
                history: props.history,
                source: 'nav',
                activation: props.activation,
                selectedSearchContextSpec: props.selectedSearchContextSpec,
                addRecentSearch,
                ...parameters,
            })
        },
        [submitSearch, props.history, props.activation, props.selectedSearchContextSpec, addRecentSearch]
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
                applySuggestionsOnEnter={applySuggestionsOnEnter}
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
                isExternalServicesUserModeAll={window.context.externalServicesUserMode === 'all'}
                structuralSearchDisabled={window.context?.experimentalFeatures?.structuralSearch === 'disabled'}
            />
        </Form>
    )
}
