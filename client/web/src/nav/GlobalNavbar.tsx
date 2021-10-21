import classNames from 'classnames'
import * as H from 'history'
import BarChartIcon from 'mdi-react/BarChartIcon'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'
import React, { useEffect, useMemo } from 'react'
import { of } from 'rxjs'
import { startWith } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { ContributableMenu } from '@sourcegraph/shared/src/api/protocol'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ActivationDropdown } from '@sourcegraph/shared/src/components/activation/ActivationDropdown'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Badge } from '@sourcegraph/web/src/components/Badge'
import { WebCommandListPopoverButton } from '@sourcegraph/web/src/components/shared'
import { FeedbackPrompt } from '@sourcegraph/web/src/nav/Feedback/FeedbackPrompt'
import { StatusMessagesNavItem } from '@sourcegraph/web/src/nav/StatusMessagesNavItem'
import { NavGroup, NavItem, NavBar, NavLink, NavActions, NavAction } from '@sourcegraph/wildcard'
import { NavDropdown } from '@sourcegraph/wildcard/src/components/NavBar/NavDropdown'

import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { BatchChangesNavItem } from '../batches/BatchChangesNavItem'
import { CodeMonitoringProps } from '../code-monitoring'
import { CodeMonitoringLogo } from '../code-monitoring/CodeMonitoringLogo'
import { BrandLogo } from '../components/branding/BrandLogo'
import { CodeInsightsProps } from '../insights/types'
import { isCodeInsightsEnabled } from '../insights/utils/is-code-insights-enabled'
import {
    KeyboardShortcutsProps,
    KEYBOARD_SHORTCUT_SHOW_COMMAND_PALETTE,
    KEYBOARD_SHORTCUT_SWITCH_THEME,
} from '../keyboardShortcuts/keyboardShortcuts'
import { LayoutRouteProps } from '../routes'
import { Settings } from '../schema/settings.schema'
import {
    PatternTypeProps,
    CaseSensitivityProps,
    OnboardingTourProps,
    ParsedSearchQueryProps,
    isSearchContextSpecAvailable,
    getGlobalSearchContextFilter,
    SearchContextInputProps,
} from '../search'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { useNavbarQueryState } from '../search/navbarSearchQueryState'
import { ThemePreferenceProps } from '../theme'
import { userExternalServicesEnabledFromTags } from '../user/settings/cloud-ga'
import { showDotComMarketing } from '../util/features'

import styles from './GlobalNavbar.module.scss'
import { ExtensionAlertAnimationProps, UserNavItem } from './UserNavItem'

interface Props
    extends SettingsCascadeProps<Settings>,
        PlatformContextProps,
        ExtensionsControllerProps,
        KeyboardShortcutsProps,
        TelemetryProps,
        ThemeProps,
        ThemePreferenceProps,
        ExtensionAlertAnimationProps,
        ActivationProps,
        Pick<ParsedSearchQueryProps, 'parsedSearchQuery'>,
        PatternTypeProps,
        CaseSensitivityProps,
        SearchContextInputProps,
        CodeMonitoringProps,
        CodeInsightsProps,
        OnboardingTourProps,
        BatchChangesProps {
    history: H.History
    location: H.Location<{ query: string }>
    authenticatedUser: AuthenticatedUser | null
    authRequired: boolean
    isSourcegraphDotCom: boolean
    showSearchBox: boolean
    routes: readonly LayoutRouteProps<{}>[]

    // Whether globbing is enabled for filters.
    globbing: boolean

    /**
     * Which variation of the global navbar to render.
     *
     * 'low-profile' renders the the navbar with no border or background. Used on the search
     * homepage.
     *
     * 'low-profile-with-logo' renders the low-profile navbar but with the homepage logo. Used on community search context pages.
     */
    variant: 'default' | 'low-profile' | 'low-profile-with-logo'

    minimalNavLinks?: boolean
    isSearchAutoFocusRequired?: boolean
    isRepositoryRelatedPage?: boolean
    branding?: typeof window.context.branding

    /** For testing only. Used because reactstrap's Popover is incompatible with react-test-renderer. */
    hideNavLinks: boolean
}

export const GlobalNavbar: React.FunctionComponent<Props> = ({
    authRequired,
    showSearchBox,
    caseSensitive,
    patternType,
    hideNavLinks,
    variant,
    isLightTheme,
    branding,
    location,
    history,
    minimalNavLinks,
    isSourcegraphDotCom,
    isRepositoryRelatedPage,
    codeInsightsEnabled,
    searchContextsEnabled,
    ...props
}) => {
    // Workaround: can't put this in optional parameter value because of https://github.com/babel/babel/issues/11166
    branding = branding ?? window.context?.branding

    const query = props.parsedSearchQuery

    const globalSearchContextSpec = useMemo(() => getGlobalSearchContextFilter(query), [query])

    // UI includes repositories section as part of the user navigation bar
    // This filter makes sure repositories feature flag is active.
    const showRepositorySection = props.authenticatedUser
        ? userExternalServicesEnabledFromTags(props.authenticatedUser.tags)
        : false

    const isSearchContextAvailable = useObservable(
        useMemo(
            () =>
                globalSearchContextSpec && searchContextsEnabled
                    ? // While we wait for the result of the `isSearchContextSpecAvailable` call, we assume the context is available
                      // to prevent flashing and moving content in the query bar. This optimizes for the most common use case where
                      // user selects a search context from the dropdown.
                      // See https://github.com/sourcegraph/sourcegraph/issues/19918 for more info.
                      isSearchContextSpecAvailable(globalSearchContextSpec.spec).pipe(startWith(true))
                    : of(false),
            [globalSearchContextSpec, searchContextsEnabled]
        )
    )

    const onNavbarQueryChange = useNavbarQueryState(state => state.setQueryState)

    useEffect(() => {
        // On a non-search related page or non-repo page, we clear the query in
        // the main query input to avoid misleading users
        // that the query is relevant in any way on those pages.
        if (!showSearchBox) {
            onNavbarQueryChange({ query: '' })
            return
        }
        // Do nothing if there is no query in the URL
        if (!query) {
            return
        }

        // If a global search context spec is available to the user, we omit it from the
        // query and move it to the search contexts dropdown
        const finalQuery =
            globalSearchContextSpec && isSearchContextAvailable && props.showSearchContext
                ? omitFilter(query, globalSearchContextSpec.filter)
                : query

        onNavbarQueryChange({ query: finalQuery })
    }, [
        showSearchBox,
        onNavbarQueryChange,
        query,
        globalSearchContextSpec,
        isSearchContextAvailable,
        props.showSearchContext,
    ])

    // CodeInsightsEnabled props controls insights appearance over OSS and Enterprise version
    // isCodeInsightsEnabled selector controls appearance based on user settings flags
    const codeInsights = props.authenticatedUser && codeInsightsEnabled && isCodeInsightsEnabled(props.settingsCascade)

    const searchNavBar = (
        <SearchNavbarItem
            {...props}
            location={location}
            history={history}
            isLightTheme={isLightTheme}
            patternType={patternType}
            caseSensitive={caseSensitive}
            isSourcegraphDotCom={isSourcegraphDotCom}
            searchContextsEnabled={searchContextsEnabled}
            isRepositoryRelatedPage={isRepositoryRelatedPage}
        />
    )

    return (
        <>
            <NavBar
                logo={
                    <BrandLogo
                        branding={branding}
                        isLightTheme={isLightTheme}
                        variant="symbol"
                        className={styles.logo}
                    />
                }
            >
                <NavGroup>
                    <NavDropdown
                        toggleItem={{ path: '/search', icon: MagnifyIcon, content: 'Code Search' }}
                        items={[
                            { path: '/search', content: 'Search home' },
                            {
                                path: '/contexts',
                                content: (
                                    <>
                                        Contexts <Badge className="ml-1" status="new" />
                                    </>
                                ),
                            },
                        ]}
                    />
                    {props.enableCodeMonitoring && (
                        <NavItem icon={CodeMonitoringLogo}>
                            <NavLink to="/code-monitoring">Monitoring</NavLink>
                        </NavItem>
                    )}
                    {/* This is the only circumstance where we show something
                         batch-changes-related even if the instance does not have batch
                         changes enabled, for marketing purposes on sourcegraph.com */}
                    {(props.batchChangesEnabled || isSourcegraphDotCom) && <BatchChangesNavItem />}
                    {codeInsights && (
                        <NavItem icon={BarChartIcon}>
                            <NavLink to="/insights/dashboards/all">Insights</NavLink>
                        </NavItem>
                    )}
                    <NavItem icon={PuzzleOutlineIcon}>
                        <NavLink to="/extensions">Extensions</NavLink>
                    </NavItem>
                    {props.activation && (
                        <NavItem>
                            <ActivationDropdown activation={props.activation} history={history} />
                        </NavItem>
                    )}
                </NavGroup>
                <NavActions>
                    {!props.authenticatedUser && (
                        <>
                            <NavAction>
                                <Link className={styles.link} to="https://about.sourcegraph.com">
                                    About <span className="d-none d-sm-inline">Sourcegraph</span>
                                </Link>
                            </NavAction>

                            {showDotComMarketing && (
                                <NavAction>
                                    <Link
                                        className={classNames('font-weight-medium', styles.link)}
                                        to="/help"
                                        target="_blank"
                                    >
                                        Docs
                                    </Link>
                                </NavAction>
                            )}
                        </>
                    )}
                    {props.authenticatedUser && (
                        <NavAction>
                            <FeedbackPrompt routes={props.routes} />
                        </NavAction>
                    )}
                    {props.authenticatedUser && (
                        <NavAction>
                            <WebCommandListPopoverButton
                                {...props}
                                location={location}
                                buttonClassName="btn btn-link p-0 m-0"
                                menu={ContributableMenu.CommandPalette}
                                keyboardShortcutForShow={KEYBOARD_SHORTCUT_SHOW_COMMAND_PALETTE}
                            />
                        </NavAction>
                    )}
                    {props.authenticatedUser &&
                        (props.authenticatedUser.siteAdmin ||
                            userExternalServicesEnabledFromTags(props.authenticatedUser.tags)) && (
                            <NavAction>
                                <StatusMessagesNavItem
                                    user={{
                                        id: props.authenticatedUser.id,
                                        username: props.authenticatedUser.username,
                                        isSiteAdmin: props.authenticatedUser?.siteAdmin || false,
                                    }}
                                    history={history}
                                />
                            </NavAction>
                        )}
                    {!props.authenticatedUser ? (
                        <>
                            <NavAction>
                                <div>
                                    <Link className="btn btn-sm btn-outline-secondary mr-1" to="/sign-in">
                                        Log in
                                    </Link>
                                    <Link className={classNames('btn btn-sm', styles.signUp)} to="/sign-up">
                                        Sign up
                                    </Link>
                                </div>
                            </NavAction>
                        </>
                    ) : (
                        <NavAction>
                            <UserNavItem
                                {...props}
                                location={location}
                                isLightTheme={isLightTheme}
                                authenticatedUser={props.authenticatedUser}
                                showDotComMarketing={showDotComMarketing}
                                showRepositorySection={showRepositorySection}
                                codeHostIntegrationMessaging={
                                    (!isErrorLike(props.settingsCascade.final) &&
                                        props.settingsCascade.final?.['alerts.codeHostIntegrationMessaging']) ||
                                    'browser-extension'
                                }
                                keyboardShortcutForSwitchTheme={KEYBOARD_SHORTCUT_SWITCH_THEME}
                            />
                        </NavAction>
                    )}
                </NavActions>
            </NavBar>
            {showSearchBox && (
                <div className="w-100 px-3 pt-2">
                    <div className="pb-2 border-bottom">{searchNavBar}</div>
                </div>
            )}
        </>
    )
}
