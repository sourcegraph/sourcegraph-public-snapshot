import * as H from 'history'
import React, { useState, useCallback, useEffect, useMemo, useRef } from 'react'
import { Form } from 'reactstrap'
import { NavbarQueryState } from 'src/stores/navbarSearchQueryState'
import shallow from 'zustand/shallow'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps, isSettingsValid } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { PatternTypeProps, ParsedSearchQueryProps, SearchContextInputProps, CaseSensitivityProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { Notices } from '../../global/Notices'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { Settings } from '../../schema/settings.schema'
import { useExperimentalFeatures, useNavbarQueryState } from '../../stores'
import { ThemePreferenceProps } from '../../theme'
import { canSubmitSearch, submitSearch, SubmitSearchParameters } from '../helpers'
import { SearchBox } from '../input/SearchBox'
import { useSearchOnboardingTour } from '../input/SearchOnboardingTour'
import { QuickLinks } from '../QuickLinks'

import styles from './SearchPageInput.module.scss'

interface Props
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        ParsedSearchQueryProps,
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sourcegraphURL'>,
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
}

const queryStateSelector = (state: NavbarQueryState): CaseSensitivityProps => ({
    caseSensitive: state.searchCaseSensitivity,
    setCaseSensitivity: state.setSearchCaseSensitivity,
})

export const SearchPageInput: React.FunctionComponent<Props> = (props: Props) => {
    /** The value entered by the user in the query input */
    const [userQueryState, setUserQueryState] = useState({
        query: props.queryPrefix ? props.queryPrefix : '',
    })
    const { caseSensitive, setCaseSensitivity } = useNavbarQueryState(queryStateSelector, shallow)
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
    const showSearchContextManagement = useExperimentalFeatures(
        features => features.showSearchContextManagement ?? false
    )
    useEffect(() => {
        setUserQueryState({ query: props.queryPrefix || '' })
    }, [props.queryPrefix])

    const quickLinks =
        (isSettingsValid<Settings>(props.settingsCascade) && props.settingsCascade.final.quicklinks) || []

    // This component is also used on the CommunitySearchContextPage.
    // The search onboarding tour should only be shown on the homepage.
    const isHomepage = useMemo(() => props.location.pathname === '/search' && !props.parsedSearchQuery, [
        props.location.pathname,
        props.parsedSearchQuery,
    ])
    const showOnboardingTour = useExperimentalFeatures(features => features.showOnboardingTour ?? false) && isHomepage

    const tourContainer = useRef<HTMLDivElement>(null)

    const { shouldFocusQueryInput, ...onboardingTourQueryInputProps } = useSearchOnboardingTour({
        ...props,
        showOnboardingTour,
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
                    patternType: props.patternType,
                    caseSensitive,
                    activation: props.activation,
                    selectedSearchContextSpec: props.selectedSearchContextSpec,
                    ...parameters,
                })
            }
        },
        [
            props.history,
            props.patternType,
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
                        caseSensitive={caseSensitive}
                        setCaseSensitivity={setCaseSensitivity}
                        submitSearchOnToggle={submitSearchOnChange}
                        queryState={userQueryState}
                        onChange={setUserQueryState}
                        onSubmit={onSubmit}
                        autoFocus={showOnboardingTour ? shouldFocusQueryInput : props.autoFocus !== false}
                    />
                </div>
                <QuickLinks quickLinks={quickLinks} className={styles.inputSubContainer} />
                <Notices className="my-3" location="home" settingsCascade={props.settingsCascade} />
            </Form>
        </div>
    )
}
