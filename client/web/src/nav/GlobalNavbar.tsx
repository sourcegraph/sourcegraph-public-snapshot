import {
    type FC,
    type MutableRefObject,
    type SetStateAction,
    useEffect,
    useLayoutEffect,
    useMemo,
    useRef,
    useState,
} from 'react'

import classNames from 'classnames'
import BarChartIcon from 'mdi-react/BarChartIcon'
import BookOutlineIcon from 'mdi-react/BookOutlineIcon'
import CommentQuoteOutline from 'mdi-react/CommentQuoteOutlineIcon'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import { type RouteObject, useLocation } from 'react-router-dom'
import useResizeObserver from 'use-resize-observer'

import { isMacPlatform } from '@sourcegraph/common'
import { shortcutDisplayName } from '@sourcegraph/shared/src/keyboardShortcuts'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import type { SearchContextInputProps } from '@sourcegraph/shared/src/search'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Button, ButtonLink, Link, ProductStatusBadge } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import type { BatchChangesProps } from '../batches'
import { BatchChangesNavItem } from '../batches/BatchChangesNavItem'
import { CodeMonitoringLogo } from '../code-monitoring/CodeMonitoringLogo'
import type { CodeMonitoringProps } from '../codeMonitoring'
import { CodyLogo } from '../cody/components/CodyLogo'
import { UpdateGlobalNav } from '../cody/update/UpdateGlobalNav'
import { BrandLogo } from '../components/branding/BrandLogo'
import { useFuzzyFinderFeatureFlags } from '../components/fuzzyFinder/FuzzyFinderFeatureFlag'
import { DeveloperSettingsGlobalNavItem } from '../devsettings/DeveloperSettingsGlobalNavItem'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { useRoutesMatch } from '../hooks'
import type { CodeInsightsProps } from '../insights/types'
import type { NotebookProps } from '../notebooks'
import { OnboardingChecklist } from '../onboarding'
import type { OwnConfigProps } from '../own/OwnConfigProps'
import { PageRoutes } from '../routes.constants'
import { isSearchJobsEnabled } from '../search-jobs/utility'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { AccessRequestsGlobalNavItem } from '../site-admin/AccessRequestsPage/AccessRequestsGlobalNavItem'
import { useDeveloperSettings, useNavbarQueryState } from '../stores'
import { eventLogger } from '../tracking/eventLogger'
import { EventName, EventLocation } from '../util/constants'

import { NavAction, NavActions, NavBar, NavGroup, NavItem, NavLink } from '.'
import { NavDropdown, type NavDropdownItem } from './NavBar/NavDropdown'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'
import { UserNavItem } from './UserNavItem'

import styles from './GlobalNavbar.module.scss'

export interface GlobalNavbarProps
    extends SettingsCascadeProps<Settings>,
        PlatformContextProps,
        TelemetryProps,
        SearchContextInputProps,
        CodeInsightsProps,
        BatchChangesProps,
        NotebookProps,
        CodeMonitoringProps,
        OwnConfigProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    isCodyApp: boolean
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
 * @param containerReference a reference to navbar container
 */
function useCalculatedNavLinkVariant(containerReference: MutableRefObject<HTMLElement | null>): 'compact' | undefined {
    const { width = 0 } = useResizeObserver({ ref: containerReference })

    const [navLinkVariant, setNavLinkVariant] = useState<'compact'>()
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
    }, [containerReference, savedWindowWidth, width])

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
    branding = window.context?.branding,
    isSourcegraphDotCom,
    isCodyApp,
    isRepositoryRelatedPage,
    codeInsightsEnabled,
    searchContextsEnabled,
    codeMonitoringEnabled,
    notebooksEnabled,
    ownEnabled,
    showFeedbackModal,
    ...props
}) => {
    const location = useLocation()

    const routeMatch = useRoutesMatch(props.routes)

    const onNavbarQueryChange = useNavbarQueryState(state => state.setQueryState)
    // Search context management is still enabled on .com
    // but should not show in the navbar. Users can still
    // access this feature via the context dropdown.
    const showSearchContext = searchContextsEnabled && !isSourcegraphDotCom
    const showCodeMonitoring = codeMonitoringEnabled && !isCodyApp && !isSourcegraphDotCom
    const showSearchNotebook = notebooksEnabled && !isCodyApp && !isSourcegraphDotCom
    const isLicensed = !!window.context?.licenseInfo || isCodyApp // Assume licensed when running as a native app
    const showBatchChanges = props.batchChangesEnabled && isLicensed && !isCodyApp && !isSourcegraphDotCom
    const [codySearchEnabled] = useFeatureFlag('cody-web-search')
    const [isAdminOnboardingEnabled] = useFeatureFlag('admin-onboarding')

    useEffect(() => {
        // On a non-search related page or non-repo page, we clear the query in
        // the main query input to avoid misleading users
        // that the query is relevant in any way on those pages.
        if (!showSearchBox) {
            onNavbarQueryChange({ query: '' })
            return
        }
    }, [showSearchBox, onNavbarQueryChange])

    const codeInsights = (codeInsightsEnabled && !isCodyApp && !isSourcegraphDotCom) ?? false

    const { fuzzyFinderNavbar } = useFuzzyFinderFeatureFlags()

    const isLightTheme = useIsLightTheme()

    const developerMode = useDeveloperSettings(settings => settings.enabled) || process.env.NODE_ENV === 'development'

    return (
        <>
            <NavBar
                logo={
                    !isCodyApp && (
                        <BrandLogo
                            branding={branding}
                            isLightTheme={isLightTheme}
                            variant="symbol"
                            className={styles.logo}
                        />
                    )
                }
            >
                <InlineNavigationPanel
                    isCodyApp={isCodyApp}
                    showSearchContext={showSearchContext}
                    showOwn={ownEnabled}
                    showCodySearch={codySearchEnabled}
                    showSearchJobs={isSearchJobsEnabled()}
                    showSearchNotebook={showSearchNotebook}
                    showCodeMonitoring={showCodeMonitoring}
                    showBatchChanges={showBatchChanges}
                    showCodeInsights={codeInsights}
                    routeMatch={routeMatch}
                    isSourcegraphDotCom={isSourcegraphDotCom}
                />

                <NavActions>
                    {developerMode && (
                        <NavAction>
                            <DeveloperSettingsGlobalNavItem />
                        </NavAction>
                    )}
                    {isCodyApp && <UpdateGlobalNav />}
                    {props.authenticatedUser?.siteAdmin && (
                        <AccessRequestsGlobalNavItem className="d-flex align-items-center py-1" />
                    )}
                    {isSourcegraphDotCom && (
                        <NavAction>
                            <Link
                                to="/get-cody"
                                className={classNames(styles.link, 'small')}
                                onClick={() => eventLogger.log(EventName.CODY_CTA, { location: EventLocation.NAV_BAR })}
                            >
                                Install Cody locally
                            </Link>
                        </NavAction>
                    )}
                    {fuzzyFinderNavbar && FuzzyFinderNavItem(props.setFuzzyFinderIsVisible)}
                    {props.authenticatedUser?.siteAdmin && !isCodyApp && (
                        <>
                            {isAdminOnboardingEnabled && (
                                <NavAction>
                                    <OnboardingChecklist />
                                </NavAction>
                            )}
                            <NavAction>
                                <StatusMessagesNavItem isCodyApp={isCodyApp} />
                            </NavAction>
                        </>
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
                                isCodyApp={isCodyApp}
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
                        isSourcegraphDotCom={isSourcegraphDotCom}
                        searchContextsEnabled={searchContextsEnabled}
                        isRepositoryRelatedPage={isRepositoryRelatedPage}
                    />
                </div>
            )}
        </>
    )
}

export interface InlineNavigationPanelProps {
    isCodyApp: boolean
    showSearchContext: boolean
    showOwn: boolean
    showCodySearch: boolean
    showSearchJobs: boolean
    showSearchNotebook: boolean
    showCodeMonitoring: boolean
    showBatchChanges: boolean
    showCodeInsights: boolean
    isSourcegraphDotCom: boolean

    /** A current react router route match */
    routeMatch?: string
    className?: string
}

export const InlineNavigationPanel: FC<InlineNavigationPanelProps> = props => {
    const {
        isCodyApp,
        showSearchContext,
        showOwn,
        showCodySearch,
        showSearchJobs,
        showSearchNotebook,
        showBatchChanges,
        showCodeInsights,
        showCodeMonitoring,
        isSourcegraphDotCom,
        routeMatch,
        className,
    } = props

    const navbarReference = useRef<HTMLDivElement | null>(null)
    const navLinkVariant = useCalculatedNavLinkVariant(navbarReference)

    const searchNavBarItems = useMemo(() => {
        const items: (NavDropdownItem | false)[] = [
            showSearchContext && { path: PageRoutes.Contexts, content: 'Contexts' },
            showOwn && { path: PageRoutes.Own, content: 'Code ownership' },
            showCodySearch && {
                path: PageRoutes.CodySearch,
                content: (
                    <>
                        Natural language search <ProductStatusBadge status="experimental" />
                    </>
                ),
            },
            showSearchJobs && {
                path: PageRoutes.SearchJobs,
                content: (
                    <>
                        Search Jobs <ProductStatusBadge className="ml-2" status="experimental" />
                    </>
                ),
            },
        ]
        return items.filter<NavDropdownItem>((item): item is NavDropdownItem => !!item)
    }, [showOwn, showSearchContext, showCodySearch, showSearchJobs])

    return (
        <NavGroup ref={navbarReference} className={classNames(className, styles.list)}>
            {!isCodyApp &&
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
                <NavLink variant={navLinkVariant} to={PageRoutes.Cody}>
                    Cody
                </NavLink>
            </NavItem>
            {showSearchNotebook && (
                <NavItem icon={BookOutlineIcon}>
                    <NavLink variant={navLinkVariant} to={PageRoutes.Notebooks}>
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
            {showCodeInsights && (
                <NavItem icon={BarChartIcon}>
                    <NavLink variant={navLinkVariant} to="/insights">
                        Insights
                    </NavLink>
                </NavItem>
            )}
            {isCodyApp && (
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
            {isSourcegraphDotCom && (
                <NavItem>
                    <NavLink variant={navLinkVariant} to="https://about.sourcegraph.com" external={true}>
                        About Sourcegraph
                    </NavLink>
                </NavItem>
            )}
        </NavGroup>
    )
}
