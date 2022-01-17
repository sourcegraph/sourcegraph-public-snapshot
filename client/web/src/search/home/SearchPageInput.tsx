import * as H from 'history'
import React, { useState, useCallback, useEffect, useRef } from 'react'
import { Form } from 'reactstrap'
import shallow from 'zustand/shallow'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps, isSettingsValid } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../../auth'
import { Notices } from '../../global/Notices'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { Settings } from '../../schema/settings.schema'
import {
    useExperimentalFeatures,
    useNavbarQueryState,
    setSearchCaseSensitivity,
    setSearchPatternType,
} from '../../stores'
import {
    NavbarQueryState,
    setSelectedSearchContext,
    submitSearchWithGlobalQueryState,
} from '../../stores/navbarSearchQueryState'
import { ThemePreferenceProps } from '../../theme'
import { fetchAutoDefinedSearchContexts, fetchSearchContexts, getUserSearchContextNamespaces } from '../backend'
import { SubmitSearchParameters } from '../helpers'
import { SearchBox } from '../input/SearchBox'
import { useSearchOnboardingTour } from '../input/SearchOnboardingTour'
import { QuickLinks } from '../QuickLinks'

import styles from './SearchPageInput.module.scss'

interface Props
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sourcegraphURL'>,
        Pick<SubmitSearchParameters, 'source'> {
    selectedSearchContextSpec?: string
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    /** Whether globbing is enabled for filters. */
    globbing: boolean
    /** A query fragment to appear at the beginning of the input. */
    queryPrefix?: string
    /** A query fragment to be prepended to queries. This will not appear in the input until a search is submitted. */
    hiddenQueryPrefix?: string
    autoFocus?: boolean
    showOnboardingTour?: boolean
}

const queryStateSelector = (
    state: NavbarQueryState
): Pick<
    NavbarQueryState,
    | 'searchCaseSensitivity'
    | 'searchPatternType'
    | 'selectedSearchContext'
    | 'defaultSearchContext'
    | 'searchContextsEnabled'
    | 'hasUserAddedRepositories'
    | 'hasUserAddedExternalServices'
    | 'hasUserSyncedPublicRepositories'
> => ({
    searchCaseSensitivity: state.searchCaseSensitivity,
    searchPatternType: state.searchPatternType,
    selectedSearchContext: state.selectedSearchContext,
    defaultSearchContext: state.defaultSearchContext,
    searchContextsEnabled: state.searchContextsEnabled,
    hasUserAddedExternalServices: state.hasUserAddedExternalServices,
    hasUserAddedRepositories: state.hasUserAddedRepositories,
    hasUserSyncedPublicRepositories: state.hasUserSyncedPublicRepositories,
})

export const SearchPageInput: React.FunctionComponent<Props> = (props: Props) => {
    /** The value entered by the user in the query input */
    const [userQueryState, setUserQueryState] = useState({
        query: props.queryPrefix ? props.queryPrefix : '',
    })
    const {
        selectedSearchContext,
        searchContextsEnabled,
        hasUserAddedExternalServices,
        hasUserAddedRepositories,
        hasUserSyncedPublicRepositories,
        defaultSearchContext,
        searchCaseSensitivity,
        searchPatternType,
    } = useNavbarQueryState(queryStateSelector, shallow)

    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
    const showSearchContextManagement = useExperimentalFeatures(
        features => features.showSearchContextManagement ?? false
    )
    useEffect(() => {
        setUserQueryState({ query: props.queryPrefix || '' })
    }, [props.queryPrefix])

    const quickLinks =
        (isSettingsValid<Settings>(props.settingsCascade) && props.settingsCascade.final.quicklinks) || []

    const tourContainer = useRef<HTMLDivElement>(null)

    const { shouldFocusQueryInput, ...onboardingTourQueryInputProps } = useSearchOnboardingTour({
        ...props,
        showOnboardingTour: props.showOnboardingTour ?? false,
        queryState: userQueryState,
        setQueryState: setUserQueryState,
        stepsContainer: tourContainer.current ?? undefined,
    })

    const submitSearchOnChange = useCallback(
        (parameters: Partial<SubmitSearchParameters> = {}) => {
            const query = props.hiddenQueryPrefix
                ? `${props.hiddenQueryPrefix} ${userQueryState.query}`
                : userQueryState.query

            submitSearchWithGlobalQueryState({
                source: 'home',
                query,
                history: props.history,
                activation: props.activation,
                selectedSearchContextSpec: props.selectedSearchContextSpec,
                ...parameters,
            })
        },
        [
            props.history,
            props.activation,
            props.selectedSearchContextSpec,
            props.hiddenQueryPrefix,
            userQueryState.query,
        ]
    )

    const onSubmit = useCallback(
        (event?: React.FormEvent): void => {
            event?.preventDefault()
            submitSearchOnChange()
        },
        [submitSearchOnChange]
    )

    return (
        <div className="d-flex flex-row flex-shrink-past-contents">
            <Form className="flex-grow-1 flex-shrink-past-contents" onSubmit={onSubmit}>
                <div data-search-page-input-container={true} className={styles.inputContainer}>
                    {/* Search onboarding tour must be rendered before the SearchBox so
                    the Monaco autocomplete suggestions are not blocked by the tour. */}
                    <div ref={tourContainer} />
                    <SearchBox
                        {...props}
                        {...onboardingTourQueryInputProps}
                        showSearchContext={showSearchContext}
                        showSearchContextManagement={showSearchContextManagement}
                        searchContextsEnabled={searchContextsEnabled}
                        selectedSearchContext={props.selectedSearchContextSpec || selectedSearchContext}
                        hasUserAddedRepositories={hasUserAddedRepositories || hasUserSyncedPublicRepositories}
                        defaultSearchContextSpec={defaultSearchContext}
                        hasUserAddedExternalServices={hasUserAddedExternalServices}
                        setSelectedSearchContextSpec={setSelectedSearchContext}
                        fetchAutoDefinedSearchContexts={fetchAutoDefinedSearchContexts}
                        fetchSearchContexts={fetchSearchContexts}
                        getUserSearchContextNamespaces={getUserSearchContextNamespaces}
                        caseSensitive={searchCaseSensitivity}
                        patternType={searchPatternType}
                        setPatternType={setSearchPatternType}
                        setCaseSensitivity={setSearchCaseSensitivity}
                        submitSearchOnToggle={submitSearchOnChange}
                        queryState={userQueryState}
                        onChange={setUserQueryState}
                        onSubmit={onSubmit}
                        autoFocus={props.showOnboardingTour ? shouldFocusQueryInput : props.autoFocus !== false}
                    />
                </div>
                <QuickLinks quickLinks={quickLinks} className={styles.inputSubContainer} />
                <Notices className="my-3" location="home" settingsCascade={props.settingsCascade} />
            </Form>
        </div>
    )
}
