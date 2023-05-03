import React, { SetStateAction, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import BarChartIcon from 'mdi-react/BarChartIcon'
import BookOutlineIcon from 'mdi-react/BookOutlineIcon'
import CommentQuoteOutline from 'mdi-react/CommentQuoteOutlineIcon'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import ShieldHalfFullIcon from 'mdi-react/ShieldHalfFullIcon'
import { RouteObject, useLocation } from 'react-router-dom'

import { isMacPlatform } from '@sourcegraph/common'
import { shortcutDisplayName } from '@sourcegraph/shared/src/keyboardShortcuts'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SearchContextInputProps } from '@sourcegraph/shared/src/search'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { addSourcegraphAppOutboundUrlParameters } from '@sourcegraph/shared/src/util/url'
import { Button, Link, ButtonLink, useWindowSize, Tooltip } from '@sourcegraph/wildcard'

import { TauriNavigation } from '../app/TauriNavigation'
import { HistoryStack } from '../app/useHistoryStack'
import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { BatchChangesNavItem } from '../batches/BatchChangesNavItem'
import { CodeMonitoringLogo } from '../code-monitoring/CodeMonitoringLogo'
import { CodeMonitoringProps } from '../codeMonitoring'
import { BrandLogo } from '../components/branding/BrandLogo'
import { useFuzzyFinderFeatureFlags } from '../components/fuzzyFinder/FuzzyFinderFeatureFlag'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { useRoutesMatch } from '../hooks'
import { CodeInsightsProps } from '../insights/types'
import { NotebookProps } from '../notebooks'
import { OwnConfigProps } from '../own/OwnConfigProps'
import { EnterprisePageRoutes, PageRoutes } from '../routes.constants'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { SentinelProps } from '../sentinel/types'
import { AccessRequestsGlobalNavItem } from '../site-admin/AccessRequestsPage/AccessRequestsGlobalNavItem'
import { useNavbarQueryState } from '../stores'
import { eventLogger } from '../tracking/eventLogger'

import { NavGroup, NavItem, NavBar, NavLink, NavActions, NavAction } from '.'
import { NavDropdown, NavDropdownItem } from './NavBar/NavDropdown'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'
import { UserNavItem } from './UserNavItem'

import styles from './GlobalNavbar.module.scss'

export interface GlobalNavbarProps
    extends SettingsCascadeProps<Settings>,
        PlatformContextProps,
        TelemetryProps,
        SearchContextInputProps,
        CodeInsightsProps,
        SentinelProps,
        BatchChangesProps,
        NotebookProps,
        CodeMonitoringProps,
        OwnConfigProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    isSourcegraphApp: boolean
    showSearchBox: boolean
    routes: RouteObject[]

    isSearchAutoFocusRequired?: boolean
    isRepositoryRelatedPage?: boolean
    branding?: typeof window.context.branding
    showKeyboardShortcutsHelp: () => void
    showFeedbackModal: () => void

    setFuzzyFinderIsVisible: React.Dispatch<SetStateAction<boolean>>

    historyStack: HistoryStack
}

/**
 * Calculates NavLink variant based whether current content fits into container or not.
 *
 * @param containerReference a reference to navbar container
 */
function useCalculatedNavLinkVariant(
    containerReference: React.MutableRefObject<HTMLDivElement | null>,
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
        // Listen for change in `authenticatedUser` to re-calculate with new dimensions,
        // based on change in navbar's content.
    }, [containerReference, savedWindowWidth, width, authenticatedUser])

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
                    {shortcutDisplayName('Mod+K')}
                </span>
            </Button>
        </NavAction>
    )
}

export const GlobalNavbar: React.FunctionComponent<React.PropsWithChildren<GlobalNavbarProps>> = ({
    showSearchBox,
    branding,
    isSourcegraphDotCom,
    isSourcegraphApp,
    isRepositoryRelatedPage,
    codeInsightsEnabled,
    sentinelEnabled,
    searchContextsEnabled,
    codeMonitoringEnabled,
    notebooksEnabled,
    ownEnabled,
    showFeedbackModal,
    historyStack,
    ...props
}) => {
    // Workaround: can't put this in optional parameter value because of https://github.com/babel/babel/issues/11166
    branding = branding ?? window.context?.branding

    const location = useLocation()

    const routeMatch = useRoutesMatch(props.routes)

    const onNavbarQueryChange = useNavbarQueryState(state => state.setQueryState)
    // Search context management is still enabled on .com
    // but should not show in the navbar. Users can still
    // access this feature via the context dropdown.
    const showSearchContext = searchContextsEnabled && !isSourcegraphDotCom
    const showCodeMonitoring = codeMonitoringEnabled
    const showSearchNotebook = notebooksEnabled

    const [isSentinelEnabled] = useFeatureFlag('sentinel')
    // TODO: Include isSourcegraphDotCom in subsequent PR
    // const showSentinel = sentinelEnabled && isSourcegraphDotCom && props.authenticatedUser?.siteAdmin
    const showSentinel = isSentinelEnabled && props.authenticatedUser?.siteAdmin

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
    const navLinkVariant = useCalculatedNavLinkVariant(navbarReference, props.authenticatedUser)

    // CodeInsightsEnabled props controls insights appearance over OSS and Enterprise version
    const codeInsights = codeInsightsEnabled

    const searchNavBarItems = useMemo(() => {
        const items: (NavDropdownItem | false)[] = [
            !!showSearchContext && { path: EnterprisePageRoutes.Contexts, content: 'Contexts' },
            ownEnabled && { path: EnterprisePageRoutes.Own, content: 'Own' },
            window.context?.codyEnabled && {
                path: EnterprisePageRoutes.CodySearch,
                content: 'Cody',
            },
        ]
        return items.filter<NavDropdownItem>((item): item is NavDropdownItem => !!item)
    }, [ownEnabled, showSearchContext])

    const { fuzzyFinderNavbar } = useFuzzyFinderFeatureFlags()

    const isLightTheme = useIsLightTheme()

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
                {isSourcegraphApp && <TauriNavigation historyStack={historyStack} />}

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
                            <NavLink variant={navLinkVariant} to="/code-monitoring">
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
                            <NavLink variant={navLinkVariant} to="/insights">
                                Insights
                            </NavLink>
                        </NavItem>
                    )}
                    {showSentinel && (
                        <NavItem icon={ShieldHalfFullIcon}>
                            <NavLink variant={navLinkVariant} to="/sentinel">
                                Sentinel
                            </NavLink>
                        </NavItem>
                    )}
                    {isSourcegraphApp && (
                        <NavDropdown
                            routeMatch="something-that-never-matches"
                            toggleItem={{
                                path: '#',
                                icon: CommentQuoteOutline,
                                content: 'Feedback',
                                variant: navLinkVariant,
                            }}
                            mobileHomeItem={{ content: 'Feedback' }}
                            items={[
                                {
                                    content: 'Join our Discord',
                                    path: 'https://about.sourcegraph.com/community',
                                    target: '_blank',
                                },
                                {
                                    content: 'File an issue',
                                    path: 'https://github.com/sourcegraph/app',
                                    target: '_blank',
                                },
                            ]}
                        />
                    )}
                </NavGroup>
                <NavActions>
                    {!props.authenticatedUser && (
                        <>
                            <NavAction>
                                <Link className={styles.link} to="https://about.sourcegraph.com">
                                    About
                                </Link>
                            </NavAction>

                            {isSourcegraphDotCom && (
                                <NavAction>
                                    <Link className={styles.link} to="/help" target="_blank">
                                        Docs
                                    </Link>
                                </NavAction>
                            )}
                        </>
                    )}
                    {isSourcegraphApp && (
                        <NavAction>
                            <Link className={styles.link} to="/app/coming-soon">
                                Coming soon
                            </Link>
                        </NavAction>
                    )}
                    {isSourcegraphApp && (
                        <ButtonLink
                            variant="secondary"
                            outline={true}
                            to={addSourcegraphAppOutboundUrlParameters(
                                'https://about.sourcegraph.com/get-started?t=enterprise',
                                'navbar'
                            )}
                            size="sm"
                            onClick={() =>
                                eventLogger.log('ClickedOnEnterpriseCTA', { location: 'NavBarSourcegraphApp' })
                            }
                        >
                            Try Sourcegraph Enterprise
                        </ButtonLink>
                    )}
                    {props.authenticatedUser?.siteAdmin && <AccessRequestsGlobalNavItem />}
                    {isSourcegraphDotCom && (
                        <>
                            <NavAction>
                                <Tooltip content="The AI code assistant you can trust to answer questions and write code for you.">
                                    <Link
                                        to="https://about.sourcegraph.com/cody"
                                        className={classNames(styles.link, 'small')}
                                        onClick={() => eventLogger.log('ClickedOnCodyCTA', { location: 'NavBar' })}
                                    >
                                        Cody
                                    </Link>
                                </Tooltip>
                            </NavAction>
                            <NavAction>
                                <Tooltip content="The Sourcegraph desktop app runs locally and works on your own private code.">
                                    <Link
                                        to="https://about.sourcegraph.com/app"
                                        className={classNames(styles.link, 'small')}
                                        onClick={() => eventLogger.log('ClickedOnAppCTA', { location: 'NavBar' })}
                                    >
                                        App
                                    </Link>
                                </Tooltip>
                            </NavAction>
                        </>
                    )}
                    {fuzzyFinderNavbar && FuzzyFinderNavItem(props.setFuzzyFinderIsVisible)}
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
                                            encodeURI(location.pathname + location.search + location.hash)
                                        }
                                        variant="secondary"
                                        outline={true}
                                        size="sm"
                                        as={Link}
                                    >
                                        Sign in
                                    </Button>
                                    {!isSourcegraphDotCom && window.context?.allowSignup && (
                                        <ButtonLink to="/sign-up" variant="primary" size="sm">
                                            Sign up
                                        </ButtonLink>
                                    )}
                                </div>
                            </NavAction>
                        </>
                    ) : (
                        <NavAction>
                            <UserNavItem
                                {...props}
                                authenticatedUser={props.authenticatedUser}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                isSourcegraphApp={isSourcegraphApp}
                                showFeedbackModal={showFeedbackModal}
                            />
                        </NavAction>
                    )}
                </NavActions>
            </NavBar>
            {showSearchBox && (
                <div className={styles.searchNavBar}>
                    <SearchNavbarItem
                        {...props}
                        isLightTheme={isLightTheme}
                        isSourcegraphDotCom={isSourcegraphDotCom}
                        searchContextsEnabled={searchContextsEnabled}
                        isRepositoryRelatedPage={isRepositoryRelatedPage}
                    />
                </div>
            )}
        </>
    )
}
