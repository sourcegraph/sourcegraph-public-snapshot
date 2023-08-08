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
import { Button, ButtonLink, Link, ProductStatusBadge, useWindowSize } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { BatchChangesProps } from '../batches'
import { BatchChangesNavItem } from '../batches/BatchChangesNavItem'
import { CodeMonitoringLogo } from '../code-monitoring/CodeMonitoringLogo'
import { CodeMonitoringProps } from '../codeMonitoring'
import { CodyLogo } from '../cody/components/CodyLogo'
import { UpdateGlobalNav } from '../cody/update/UpdateGlobalNav'
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

import { NavAction, NavActions, NavBar, NavGroup, NavItem, NavLink } from '.'
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
    const showCodeMonitoring = codeMonitoringEnabled && !isSourcegraphApp && !isSourcegraphDotCom
    const showSearchNotebook = notebooksEnabled && !isSourcegraphApp && !isSourcegraphDotCom
    const isLicensed = !!window.context?.licenseInfo || isSourcegraphApp // Assume licensed when running as a native app
    const showBatchChanges = props.batchChangesEnabled && isLicensed && !isSourcegraphApp && !isSourcegraphDotCom
    const [codySearchEnabled] = useFeatureFlag('cody-web-search')

    const [isSentinelEnabled] = useFeatureFlag('sentinel')
    // TODO: Include isSourcegraphDotCom in subsequent PR
    // const showSentinel = sentinelEnabled && isSourcegraphDotCom && props.authenticatedUser?.siteAdmin
    const showSentinel = isSentinelEnabled && props.authenticatedUser?.siteAdmin && !isSourcegraphApp

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
    const codeInsights = codeInsightsEnabled && !isSourcegraphApp && !isSourcegraphDotCom

    const searchNavBarItems = useMemo(() => {
        const items: (NavDropdownItem | false)[] = [
            !!showSearchContext && { path: EnterprisePageRoutes.Contexts, content: 'Contexts' },
            ownEnabled && { path: EnterprisePageRoutes.Own, content: 'Code ownership' },
            codySearchEnabled && {
                path: EnterprisePageRoutes.CodySearch,
                content: (
                    <>
                        Natural language search <ProductStatusBadge status="experimental" />
                    </>
                ),
            },
        ]
        return items.filter<NavDropdownItem>((item): item is NavDropdownItem => !!item)
    }, [ownEnabled, showSearchContext, codySearchEnabled])

    const { fuzzyFinderNavbar } = useFuzzyFinderFeatureFlags()

    const isLightTheme = useIsLightTheme()

    return (
        <>
            <NavBar
                ref={navbarReference}
                logo={
                    !isSourcegraphApp && (
                        <BrandLogo
                            branding={branding}
                            isLightTheme={isLightTheme}
                            variant="symbol"
                            className={styles.logo}
                        />
                    )
                }
            >
                <NavGroup>
                    {!isSourcegraphApp &&
                        (searchNavBarItems.length > 0 ? (
                            <NavDropdown
                                toggleItem={{
                                    path: PageRoutes.Search,
                                    altPath: PageRoutes.RepoContainer,
                                    icon: MagnifyIcon,
                                    content: 'Code Search',
                                    variant: navLinkVariant,
                                }}
                                routeMatch={routeMatch}
                                homeItem={{ content: 'Search home' }}
                                items={searchNavBarItems}
                                name="search"
                            />
                        ) : (
                            <NavItem icon={MagnifyIcon}>
                                <NavLink variant={navLinkVariant} to={PageRoutes.Search}>
                                    Code Search
                                </NavLink>
                            </NavItem>
                        ))}
                    <NavItem icon={CodyLogo}>
                        <NavLink variant={navLinkVariant} to={EnterprisePageRoutes.Cody}>
                            Cody AI
                        </NavLink>
                    </NavItem>
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
                    {showBatchChanges && <BatchChangesNavItem variant={navLinkVariant} />}
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
                            items={[
                                {
                                    content: 'Join our Discord',
                                    path: 'https://discord.com/servers/sourcegraph-969688426372825169',
                                    target: '_blank',
                                },
                                {
                                    content: 'File an issue',
                                    path: 'https://github.com/sourcegraph/app',
                                    target: '_blank',
                                },
                            ]}
                            name="feedback"
                        />
                    )}
                </NavGroup>
                <NavActions>
                    {isSourcegraphApp && <UpdateGlobalNav />}
                    {!props.authenticatedUser && !isSourcegraphDotCom && (
                        <NavAction>
                            <Link className={styles.link} to="https://about.sourcegraph.com">
                                About Sourcegraph
                            </Link>
                        </NavAction>
                    )}
                    {props.authenticatedUser?.siteAdmin && <AccessRequestsGlobalNavItem />}
                    {isSourcegraphDotCom && (
                        <NavAction>
                            <Link
                                to="/get-cody"
                                className={classNames(styles.link, 'small')}
                                onClick={() => eventLogger.log('ClickedOnCodyCTA', { location: 'NavBar' })}
                            >
                                Install Cody locally
                            </Link>
                        </NavAction>
                    )}
                    {fuzzyFinderNavbar && FuzzyFinderNavItem(props.setFuzzyFinderIsVisible)}
                    {props.authenticatedUser?.siteAdmin && !isSourcegraphApp && (
                        <NavAction>
                            <StatusMessagesNavItem isSourcegraphApp={isSourcegraphApp} />
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
