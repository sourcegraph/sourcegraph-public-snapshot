import React, { useState, useCallback, useEffect, useRef } from 'react'

import * as H from 'history'
import { NavbarQueryState } from 'src/stores/navbarSearchQueryState'
import shallow from 'zustand/shallow'

import { Form } from '@sourcegraph/branded/src/components/Form'
import {
    SearchContextInputProps,
    CaseSensitivityProps,
    SearchPatternTypeProps,
    SubmitSearchParameters,
    canSubmitSearch,
} from '@sourcegraph/search'
import { SearchBox } from '@sourcegraph/search-ui'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { KeyboardShortcutsProps } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps, isSettingsValid } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../../auth'
import { Notices } from '../../global/Notices'
import {
    useExperimentalFeatures,
    useNavbarQueryState,
    setSearchCaseSensitivity,
    setSearchPatternType,
} from '../../stores'
import { ThemePreferenceProps } from '../../theme'
import { submitSearch } from '../helpers'
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
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sourcegraphURL' | 'requestGraphQL'>,
        Pick<SubmitSearchParameters, 'source'>,
        SearchContextInputProps {
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
): Pick<CaseSensitivityProps, 'caseSensitive'> & SearchPatternTypeProps => ({
    caseSensitive: state.searchCaseSensitivity,
    patternType: state.searchPatternType,
})

export const SearchPageInput: React.FunctionComponent<React.PropsWithChildren<Props>> = (props: Props) => {
    /** The value entered by the user in the query input */
    const [userQueryState, setUserQueryState] = useState({
        query: props.queryPrefix ? props.queryPrefix : '',
    })
    const { caseSensitive, patternType } = useNavbarQueryState(queryStateSelector, shallow)
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
    const showSearchContextManagement = useExperimentalFeatures(
        features => features.showSearchContextManagement ?? false
    )
    const editorComponent = useExperimentalFeatures(features => features.editor ?? 'monaco')

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

            if (canSubmitSearch(query, props.selectedSearchContextSpec)) {
                submitSearch({
                    source: 'home',
                    query,
                    history: props.history,
                    patternType,
                    caseSensitive,
                    activation: props.activation,
                    selectedSearchContextSpec: props.selectedSearchContextSpec,
                    ...parameters,
                })
            }
        },
        [
            props.history,
            patternType,
            caseSensitive,
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

    // We want to prevent autofocus by default on devices with touch as their only input method.
    // Touch only devices result in the onscreen keyboard not showing until the input loses focus and
    // gets focused again by the user. The logic is not fool proof, but should rule out majority of cases
    // where a touch enabled device has a physical keyboard by relying on detection of a fine pointer with hover ability.
    const isTouchOnlyDevice =
        !window.matchMedia('(any-pointer:fine)').matches && window.matchMedia('(any-hover:none)').matches

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
                        editorComponent={editorComponent}
                        showSearchContext={showSearchContext}
                        showSearchContextManagement={showSearchContextManagement}
                        caseSensitive={caseSensitive}
                        patternType={patternType}
                        setPatternType={setSearchPatternType}
                        setCaseSensitivity={setSearchCaseSensitivity}
                        submitSearchOnToggle={submitSearchOnChange}
                        queryState={userQueryState}
                        onChange={setUserQueryState}
                        onSubmit={onSubmit}
                        autoFocus={
                            props.showOnboardingTour
                                ? shouldFocusQueryInput
                                : !isTouchOnlyDevice && props.autoFocus !== false
                        }
                        isExternalServicesUserModeAll={window.context.externalServicesUserMode === 'all'}
                        structuralSearchDisabled={window.context?.experimentalFeatures?.structuralSearch === 'disabled'}
                    />
                </div>
                <QuickLinks quickLinks={quickLinks} className={styles.inputSubContainer} />
                <Notices className="my-3" location="home" settingsCascade={props.settingsCascade} />
            </Form>
        </div>
    )
}
