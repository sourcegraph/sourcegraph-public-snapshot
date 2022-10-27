import React, { SetStateAction, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import BarChartIcon from 'mdi-react/BarChartIcon'
import BookOutlineIcon from 'mdi-react/BookOutlineIcon'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import PuzzleOutlineIcon from 'mdi-react/PuzzleOutlineIcon'

import { ContributableMenu } from '@sourcegraph/client-api'
import { isErrorLike, isMacPlatform } from '@sourcegraph/common'
import { SearchContextInputProps } from '@sourcegraph/search'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { Button, Link, ButtonLink, useWindowSize, FeedbackPrompt, PopoverTrigger } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { BatchChangesNavItem } from '../batches/BatchChangesNavItem'
import { CodeMonitoringLogo } from '../code-monitoring/CodeMonitoringLogo'
import { CodeMonitoringProps } from '../codeMonitoring'
import { ActivationDropdown } from '../components/ActivationDropdown'
import { BrandLogo } from '../components/branding/BrandLogo'
import { getFuzzyFinderFeatureFlags } from '../components/fuzzyFinder/FuzzyFinderFeatureFlag'
import { renderShortcutKey } from '../components/KeyboardShortcutsHelp/KeyboardShortcutsHelp'
import { WebCommandListPopoverButton } from '../components/shared'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { useHandleSubmitFeedback, useRoutesMatch } from '../hooks'
import { CodeInsightsProps } from '../insights/types'
import { isCodeInsightsEnabled } from '../insights/utils/is-code-insights-enabled'
import { NotebookProps } from '../notebooks'
import { LayoutRouteProps } from '../routes'
import { EnterprisePageRoutes, PageRoutes } from '../routes.constants'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { useExperimentalFeatures, useNavbarQueryState } from '../stores'
import { ThemePreferenceProps } from '../theme'
import { eventLogger } from '../tracking/eventLogger'
import { showDotComMarketing } from '../util/features'

import { NavDropdown, NavDropdownItem } from './NavBar/NavDropdown'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'
import { UserNavItem } from './UserNavItem'

import { NavGroup, NavItem, NavBar, NavLink, NavActions, NavAction } from '.'

import styles from './GlobalNavbar.module.scss'

export interface GlobalNavbarProps
    extends SettingsCascadeProps<Settings>,
        PlatformContextProps,
        ExtensionsControllerProps,
        TelemetryProps,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        SearchContextInputProps,
        CodeInsightsProps,
        BatchChangesProps,
        NotebookProps,
        CodeMonitoringProps {
    history: H.History
    location: H.Location
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    showSearchBox: boolean
    routes: readonly LayoutRouteProps<{}>[]

    // Whether globbing is enabled for filters.
    globbing: boolean
    isSearchAutoFocusRequired?: boolean
    isRepositoryRelatedPage?: boolean
    enableLegacyExtensions?: boolean
    branding?: typeof window.context.branding
    showKeyboardShortcutsHelp: () => void

    setFuzzyFinderIsVisible: React.Dispatch<SetStateAction<boolean>>
}

/**
 * Calculates NavLink variant based whether current content fits into container or not.
 *
 * @param containerReference a reference to navbar container
 */
function useCalculatedNavLinkVariant(
    containerReference: React.MutableRefObject<HTMLDivElement | null>,
    activation: GlobalNavbarProps['activation'],
    authenticatedUser: GlobalNavbarProps['authenticatedUser']
): 'compact' | undefined {
    const [navLinkVariant, setNavLinkVariant] = useState<'compact'>()
    const { width } = useWindowSize()
    const [savedWindowWidth, setSavedWindowWidth] = useState<number>()

    useLayoutEffect(() => {
        const container = containerReference.current
        if (!container) {
            return
        }
        if (container.offsetWidth < container.scrollWidth) {
            setNavLinkVariant('compact')
            setSavedWindowWidth(width)
        } else if (savedWindowWidth && width > savedWindowWidth) {
            setNavLinkVariant(undefined)
        }
        // Listen for change in `authenticatedUser` and `activation` to re-calculate with new dimensions,
        // based on change in navbar's content.
    }, [containerReference, savedWindowWidth, width, authenticatedUser, activation])

    return navLinkVariant
}

function FuzzyFinderNavItem(setFuzzyFinderVisible: React.Dispatch<SetStateAction<boolean>>): JSX.Element {
    return (
        <NavAction className="d-none d-sm-flex">
            <Button
                onClick={() => setFuzzyFinderVisible(true)}
                className={classNames(styles.fuzzyFinderItem)}
                size="sm"
            >
                <span aria-hidden={true} aria-label={isMacPlatform() ? 'command-k' : 'ctrl-k'}>
                    {isMacPlatform() ? renderShortcutKey('Mod') + 'K' : [renderShortcutKey('Mod'), 'K'].join('+')}
                </span>
            </Button>
        </NavAction>
    )
}
const AnalyticsNavItem: React.FunctionComponent = () => {
    const [isAdminAnalyticsDisabled] = useFeatureFlag('admin-analytics-disabled', false)

    if (isAdminAnalyticsDisabled) {
        return null
    }

    return (
        <NavAction className="d-none d-sm-flex">
            <Link to="/site-admin" className={styles.link}>
                Analytics
            </Link>
        </NavAction>
    )
}

export const GlobalNavbar: React.FunctionComponent<React.PropsWithChildren<GlobalNavbarProps>> = ({
    showSearchBox,
    isLightTheme,
    branding,
    location,
    history,
    isSourcegraphDotCom,
    isRepositoryRelatedPage,
    codeInsightsEnabled,
    searchContextsEnabled,
    codeMonitoringEnabled,
    notebooksEnabled,
    extensionsController,
    enableLegacyExtensions,
    ...props
}) => {
    // Workaround: can't put this in optional parameter value because of https://github.com/babel/babel/issues/11166
    branding = branding ?? window.context?.branding

    const routeMatch = useRoutesMatch(props.routes)
    const { handleSubmitFeedback } = useHandleSubmitFeedback({
        routeMatch,
    })

    const onNavbarQueryChange = useNavbarQueryState(state => state.setQueryState)
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext) && searchContextsEnabled
    const showCodeMonitoring = useExperimentalFeatures(features => features.codeMonitoring) && codeMonitoringEnabled
    const showSearchNotebook = useExperimentalFeatures(features => features.showSearchNotebook) && notebooksEnabled

    useEffect(() => {
        // On a non-search related page or non-repo page, we clear the query in
        // the main query input to avoid misleading users
        // that the query is relevant in any way on those pages.
        if (!showSearchBox) {
            onNavbarQueryChange({ query: '' })
            return
        }
    }, [showSearchBox, onNavbarQueryChange])

    const navbarReference = useRef<HTMLDivElement | null>(null)
    const navLinkVariant = useCalculatedNavLinkVariant(navbarReference, props.activation, props.authenticatedUser)

    // CodeInsightsEnabled props controls insights appearance over OSS and Enterprise version
    // isCodeInsightsEnabled selector controls appearance based on user settings flags
    const codeInsights = codeInsightsEnabled && isCodeInsightsEnabled(props.settingsCascade)

    const searchNavBarItems = useMemo(() => {
        const items: (NavDropdownItem | false)[] = [
            !!showSearchContext && { path: EnterprisePageRoutes.Contexts, content: 'Contexts' },
        ]
        return items.filter<NavDropdownItem>((item): item is NavDropdownItem => !!item)
    }, [showSearchContext])

    const { fuzzyFinderNavbar } = getFuzzyFinderFeatureFlags(props.settingsCascade.final) ?? false

    return (
        <>
            <NavBar
                ref={navbarReference}
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
                    {searchNavBarItems.length > 0 ? (
                        <NavDropdown
                            toggleItem={{
                                path: PageRoutes.Search,
                                altPath: PageRoutes.RepoContainer,
                                icon: MagnifyIcon,
                                content: 'Code Search',
                                variant: navLinkVariant,
                            }}
                            routeMatch={routeMatch}
                            mobileHomeItem={{ content: 'Search home' }}
                            items={searchNavBarItems}
                        />
                    ) : (
                        <NavItem icon={MagnifyIcon}>
                            <NavLink variant={navLinkVariant} to={PageRoutes.Search}>
                                Code Search
                            </NavLink>
                        </NavItem>
                    )}
                    {showSearchNotebook && (
                        <NavItem icon={BookOutlineIcon}>
                            <NavLink variant={navLinkVariant} to={EnterprisePageRoutes.Notebooks}>
                                Notebooks
                            </NavLink>
                        </NavItem>
                    )}
                    {showCodeMonitoring && (
                        <NavItem icon={CodeMonitoringLogo}>
                            <NavLink variant={navLinkVariant} to={EnterprisePageRoutes.CodeMonitoring}>
                                Monitoring
                            </NavLink>
                        </NavItem>
                    )}
                    {/* This is the only circumstance where we show something
                         batch-changes-related even if the instance does not have batch
                         changes enabled, for marketing purposes on sourcegraph.com */}
                    {(props.batchChangesEnabled || isSourcegraphDotCom) && (
                        <BatchChangesNavItem variant={navLinkVariant} />
                    )}
                    {codeInsights && (
                        <NavItem icon={BarChartIcon}>
                            <NavLink variant={navLinkVariant} to={EnterprisePageRoutes.Insights}>
                                Insights
                            </NavLink>
                        </NavItem>
                    )}
                    {enableLegacyExtensions && (
                        <NavItem icon={PuzzleOutlineIcon}>
                            <NavLink variant={navLinkVariant} to={PageRoutes.Extensions}>
                                Extensions
                            </NavLink>
                        </NavItem>
                    )}
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
                    {props.authenticatedUser?.siteAdmin && <AnalyticsNavItem />}
                    {fuzzyFinderNavbar && FuzzyFinderNavItem(props.setFuzzyFinderIsVisible)}
                    {props.authenticatedUser && (
                        <NavAction>
                            <FeedbackPrompt onSubmit={handleSubmitFeedback} productResearchEnabled={true}>
                                <PopoverTrigger
                                    as={Button}
                                    aria-label="Feedback"
                                    variant="secondary"
                                    outline={true}
                                    size="sm"
                                    className={styles.feedbackTrigger}
                                >
                                    <span>Feedback</span>
                                </PopoverTrigger>
                            </FeedbackPrompt>
                        </NavAction>
                    )}
                    {props.authenticatedUser && extensionsController !== null && enableLegacyExtensions && (
                        <NavAction>
                            <WebCommandListPopoverButton
                                {...props}
                                extensionsController={extensionsController}
                                location={location}
                                menu={ContributableMenu.CommandPalette}
                            />
                        </NavAction>
                    )}
                    {props.authenticatedUser?.siteAdmin && (
                        <NavAction>
                            <StatusMessagesNavItem />
                        </NavAction>
                    )}
                    {!props.authenticatedUser ? (
                        <>
                            <NavAction>
                                <div>
                                    <Button
                                        className="mr-1"
                                        to={
                                            '/sign-in?returnTo=' +
                                            encodeURI(history.location.pathname + history.location.search)
                                        }
                                        variant="secondary"
                                        outline={true}
                                        size="sm"
                                        as={Link}
                                    >
                                        Sign in
                                    </Button>
                                    <ButtonLink
                                        className={styles.signUp}
                                        to={buildGetStartedURL('nav')}
                                        size="sm"
                                        onClick={() => eventLogger.log('ClickedOnTopNavTrialButton')}
                                    >
                                        Get free trial
                                    </ButtonLink>
                                </div>
                            </NavAction>
                        </>
                    ) : (
                        <NavAction>
                            <UserNavItem
                                {...props}
                                isLightTheme={isLightTheme}
                                authenticatedUser={props.authenticatedUser}
                                showDotComMarketing={showDotComMarketing}
                                codeHostIntegrationMessaging={
                                    (!isErrorLike(props.settingsCascade.final) &&
                                        props.settingsCascade.final?.['alerts.codeHostIntegrationMessaging']) ||
                                    'browser-extension'
                                }
                            />
                        </NavAction>
                    )}
                </NavActions>
            </NavBar>
            {showSearchBox && (
                <div className="w-100 px-3 pt-2">
                    <div className="pb-2 border-bottom">
                        <SearchNavbarItem
                            {...props}
                            location={location}
                            history={history}
                            isLightTheme={isLightTheme}
                            isSourcegraphDotCom={isSourcegraphDotCom}
                            searchContextsEnabled={searchContextsEnabled}
                            isRepositoryRelatedPage={isRepositoryRelatedPage}
                        />
                    </div>
                </div>
            )}
        </>
    )
}
