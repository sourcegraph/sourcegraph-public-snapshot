import * as H from 'history'
import React, { useEffect, useState, useCallback } from 'react'
import {
    parseSearchURLQuery,
    PatternTypeProps,
    InteractiveSearchProps,
    CaseSensitivityProps,
    SmartSearchFieldProps,
    CopyQueryButtonProps,
} from '..'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { Notices } from '../../global/Notices'
import { Settings } from '../../schema/settings.schema'
import { ThemeProps } from '../../../../shared/src/theme'
import { eventLogger, EventLoggerProps } from '../../tracking/eventLogger'
import { ThemePreferenceProps } from '../../theme'
import { limitString } from '../../util'
import { submitSearch } from '../helpers'
import { QuickLinks } from '../QuickLinks'
import { QueryInput } from './QueryInput'
import { LazyMonacoQueryInput } from './LazyMonacoQueryInput'
import { SearchButton } from './SearchButton'
import { SearchScopes } from './SearchScopes'
import { InteractiveModeInput } from './interactive/InteractiveModeInput'
import { KeyboardShortcutsProps, KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '../../keyboardShortcuts/keyboardShortcuts'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SearchModeToggle } from './interactive/SearchModeToggle'
import { Link } from '../../../../shared/src/components/Link'
import { BrandLogo } from '../../components/branding/BrandLogo'
import { VersionContextDropdown } from '../../nav/VersionContextDropdown'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { VersionContext } from '../../schema/site.schema'

interface Props
    extends SettingsCascadeProps,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        KeyboardShortcutsProps,
        EventLoggerProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        InteractiveSearchProps,
        SmartSearchFieldProps,
        CopyQueryButtonProps,
        VersionContextProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
    setVersionContext: (versionContext: string | undefined) => void
    availableVersionContexts: VersionContext[] | undefined

    // For NavLinks
    authRequired?: boolean
    showCampaigns: boolean
}

/**
 * The search page
 */
export const SearchPage: React.FunctionComponent<Props> = props => {
    useEffect(() => eventLogger.logViewEvent('Home'))

    const queryFromUrl = parseSearchURLQuery(props.location.search) || ''

    /** The query cursor position and value entered by the user in the query input */
    const [userQueryState, setUserQueryState] = useState({
        query: queryFromUrl,
        cursorPosition: queryFromUrl.length,
    })

    const quickLinks =
        (isSettingsValid<Settings>(props.settingsCascade) && props.settingsCascade.final.quicklinks) || []

    const onSubmit = useCallback(
        (event?: React.FormEvent<HTMLFormElement>): void => {
            // False positive
            // eslint-disable-next-line no-unused-expressions
            event?.preventDefault()
            submitSearch({ ...props, query: userQueryState.query, source: 'home' })
        },
        [props, userQueryState.query]
    )

    const pageTitle = queryFromUrl ? `${limitString(userQueryState.query, 25, true)}` : undefined

    return (
        <div className="search-page">
            <PageTitle title={pageTitle} />
            <BrandLogo className="search-page__logo" isLightTheme={props.isLightTheme} />
            <div className="search-page__container">
                <div className="d-flex flex-row flex-shrink-past-contents">
                    {props.splitSearchModes && props.interactiveSearchMode ? (
                        <InteractiveModeInput
                            {...props}
                            navbarSearchState={userQueryState}
                            onNavbarQueryChange={setUserQueryState}
                            toggleSearchMode={props.toggleSearchMode}
                            lowProfile={false}
                        />
                    ) : (
                        <>
                            <Form className="flex-grow-1 flex-shrink-past-contents" onSubmit={onSubmit}>
                                <div className="search-page__input-container">
                                    {props.splitSearchModes && (
                                        <SearchModeToggle
                                            {...props}
                                            interactiveSearchMode={props.interactiveSearchMode}
                                        />
                                    )}
                                    <VersionContextDropdown
                                        history={props.history}
                                        caseSensitive={props.caseSensitive}
                                        patternType={props.patternType}
                                        navbarSearchQuery={userQueryState.query}
                                        versionContext={props.versionContext}
                                        setVersionContext={props.setVersionContext}
                                        availableVersionContexts={props.availableVersionContexts}
                                    />
                                    {props.smartSearchField ? (
                                        <LazyMonacoQueryInput
                                            {...props}
                                            hasGlobalQueryBehavior={true}
                                            queryState={userQueryState}
                                            onChange={setUserQueryState}
                                            onSubmit={onSubmit}
                                            autoFocus={true}
                                        />
                                    ) : (
                                        <QueryInput
                                            {...props}
                                            value={userQueryState}
                                            onChange={setUserQueryState}
                                            autoFocus="cursor-at-end"
                                            hasGlobalQueryBehavior={true}
                                            patternType={props.patternType}
                                            setPatternType={props.setPatternType}
                                            withSearchModeToggle={props.splitSearchModes}
                                            keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR}
                                        />
                                    )}
                                    <SearchButton />
                                </div>
                                <div className="search-page__input-sub-container">
                                    {!props.splitSearchModes && (
                                        <Link className="btn btn-link btn-sm pl-0" to="/search/query-builder">
                                            Query builder
                                        </Link>
                                    )}
                                    <SearchScopes
                                        history={props.history}
                                        query={userQueryState.query}
                                        authenticatedUser={props.authenticatedUser}
                                        settingsCascade={props.settingsCascade}
                                        patternType={props.patternType}
                                        versionContext={props.versionContext}
                                    />
                                </div>
                                <QuickLinks quickLinks={quickLinks} className="search-page__input-sub-container" />
                                <Notices
                                    className="my-3"
                                    location="home"
                                    settingsCascade={props.settingsCascade}
                                    history={props.history}
                                />
                            </Form>
                        </>
                    )}
                </div>
            </div>
        </div>
    )
}
