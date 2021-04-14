import * as H from 'history'
import React, { useState, useCallback, useEffect, useMemo } from 'react'
import { Form } from 'reactstrap'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps, isSettingsValid } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import {
    PatternTypeProps,
    CaseSensitivityProps,
    CopyQueryButtonProps,
    OnboardingTourProps,
    ParsedSearchQueryProps,
    SearchContextProps,
} from '..'
import { AuthenticatedUser } from '../../auth'
import { Notices } from '../../global/Notices'
import { KeyboardShortcutsProps } from '../../keyboardShortcuts/keyboardShortcuts'
import { VersionContextDropdown } from '../../nav/VersionContextDropdown'
import { Settings } from '../../schema/settings.schema'
import { VersionContext } from '../../schema/site.schema'
import { ThemePreferenceProps } from '../../theme'
import { submitSearch, SubmitSearchParameters } from '../helpers'
import { QuickLinks } from '../QuickLinks'

import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { SearchButton } from './SearchButton'
import { useSearchOnboardingTour } from './SearchOnboardingTour'

interface Props
    extends SettingsCascadeProps<Settings>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'sourcegraphURL'>,
        CopyQueryButtonProps,
        Pick<SubmitSearchParameters, 'source'>,
        VersionContextProps,
        SearchContextProps,
        OnboardingTourProps {
    authenticatedUser: AuthenticatedUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    setVersionContext: (versionContext: string | undefined) => Promise<void>
    availableVersionContexts: VersionContext[] | undefined
    /** Whether globbing is enabled for filters. */
    globbing: boolean
    // Whether to additionally highlight or provide hovers for tokens, e.g., regexp character sets.
    enableSmartQuery: boolean
    /** Show the query builder link. */
    showQueryBuilder: boolean
    /** A query fragment to appear at the beginning of the input. */
    queryPrefix?: string
    /** A query fragment to be prepended to queries. This will not appear in the input until a search is submitted. */
    hiddenQueryPrefix?: string
    /** Don't show the version contexts dropdown. */
    hideVersionContexts?: boolean
    autoFocus?: boolean
}

export const SearchPageInput: React.FunctionComponent<Props> = (props: Props) => {
    /** The value entered by the user in the query input */
    const [userQueryState, setUserQueryState] = useState({
        query: props.queryPrefix ? props.queryPrefix : '',
    })

    useEffect(() => {
        setUserQueryState({ query: props.queryPrefix || '' })
    }, [props.queryPrefix])

    const quickLinks =
        (isSettingsValid<Settings>(props.settingsCascade) && props.settingsCascade.final.quicklinks) || []

    // This component is also used on the RepogroupPage.
    // The search onboarding tour should only be shown on the homepage.
    const isHomepage = useMemo(() => props.location.pathname === '/search' && !props.parsedSearchQuery, [
        props.location.pathname,
        props.parsedSearchQuery,
    ])
    const showOnboardingTour = props.showOnboardingTour && isHomepage

    const {
        additionalQueryParameters,
        shouldFocusQueryInput,
        ...onboardingTourQueryInputProps
    } = useSearchOnboardingTour({
        ...props,
        showOnboardingTour,
        inputLocation: 'search-homepage',
        queryState: userQueryState,
        setQueryState: setUserQueryState,
    })
    const onSubmit = useCallback(
        (event?: React.FormEvent<HTMLFormElement>): void => {
            event?.preventDefault()
            submitSearch({
                ...props,
                query: props.hiddenQueryPrefix
                    ? `${props.hiddenQueryPrefix} ${userQueryState.query}`
                    : userQueryState.query,
                source: 'home',
                searchParameters: additionalQueryParameters,
            })
        },
        [props, userQueryState.query, additionalQueryParameters]
    )

    return (
        <div className="d-flex flex-row flex-shrink-past-contents">
            <Form className="flex-grow-1 flex-shrink-past-contents" onSubmit={onSubmit}>
                <div className="search-page__input-container">
                    {!props.hideVersionContexts && (
                        <VersionContextDropdown
                            history={props.history}
                            caseSensitive={props.caseSensitive}
                            patternType={props.patternType}
                            navbarSearchQuery={userQueryState.query}
                            versionContext={props.versionContext}
                            setVersionContext={props.setVersionContext}
                            availableVersionContexts={props.availableVersionContexts}
                            selectedSearchContextSpec={props.selectedSearchContextSpec}
                        />
                    )}
                    <LazyMonacoQueryInput
                        {...props}
                        {...onboardingTourQueryInputProps}
                        submitSearchOnSearchContextChange={false}
                        hasGlobalQueryBehavior={true}
                        queryState={userQueryState}
                        onChange={setUserQueryState}
                        onSubmit={onSubmit}
                        autoFocus={showOnboardingTour ? shouldFocusQueryInput : props.autoFocus !== false}
                        showSearchContextHighlightTourStep={true}
                    />
                    <SearchButton />
                </div>
                {props.showQueryBuilder && (
                    <div className="search-page__input-sub-container">
                        <Link className="btn btn-link btn-sm pl-0" to="/search/query-builder">
                            Query builder
                        </Link>
                    </div>
                )}
                <QuickLinks quickLinks={quickLinks} className="search-page__input-sub-container" />
                <Notices
                    className="my-3"
                    location="home"
                    settingsCascade={props.settingsCascade}
                    history={props.history}
                />
            </Form>
        </div>
    )
}
