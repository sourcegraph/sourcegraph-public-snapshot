import * as H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { authRequired } from '../auth'
import {
    parseSearchURLQuery,
    PatternTypeProps,
    InteractiveSearchProps,
    CaseSensitivityProps,
    SmartSearchFieldProps,
} from '../search'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { EventLoggerProps } from '../tracking/eventLogger'
import { showDotComMarketing } from '../util/features'
import { NavLinks } from './NavLinks'
import { ThemeProps } from '../../../shared/src/theme'
import { ThemePreferenceProps } from '../theme'
import { KeyboardShortcutsProps } from '../keyboardShortcuts/keyboardShortcuts'
import { QueryState } from '../search/helpers'
import { InteractiveModeInput } from '../search/input/interactive/InteractiveModeInput'
import { FiltersToTypeAndValue } from '../../../shared/src/search/interactive/util'
import { SearchModeToggle } from '../search/input/interactive/SearchModeToggle'
import { Link } from '../../../shared/src/components/Link'
import { convertPlainTextToInteractiveQuery } from '../search/input/helpers'

interface Props
    extends SettingsCascadeProps,
        PlatformContextProps,
        ExtensionsControllerProps,
        KeyboardShortcutsProps,
        EventLoggerProps,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps,
        PatternTypeProps,
        CaseSensitivityProps,
        InteractiveSearchProps,
        SmartSearchFieldProps {
    history: H.History
    location: H.Location<{ query: string }>
    authenticatedUser: GQL.IUser | null
    navbarSearchQueryState: QueryState
    onNavbarQueryChange: (queryState: QueryState) => void
    isSourcegraphDotCom: boolean
    isSearchRelatedPage: boolean
    showCampaigns: boolean

    /**
     * Whether to hide the global search input. Use this when the page has a search input that would
     * conflict with or be confusing with the global search input.
     */
    hideGlobalSearchInput: boolean

    /**
     * Whether to use the low-profile form of the navbar, which has no border or background. Used on the search
     * homepage.
     */
    lowProfile: boolean

    filtersInQuery: FiltersToTypeAndValue
    splitSearchModes: boolean
    interactiveSearchMode: boolean
    toggleSearchMode: (event: React.MouseEvent<HTMLAnchorElement>) => void

    /** For testing only. Used because reactstrap's Popover is incompatible with react-test-renderer. */
    hideNavLinks: boolean
}

interface State {
    authRequired?: boolean
}

export class GlobalNavbar extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        // In interactive search mode, the InteractiveModeInput component will handle updating the inputs.
        if (!props.interactiveSearchMode) {
            // Reads initial state from the props (i.e. URL parameters).
            const query = parseSearchURLQuery(props.location.search || '')
            if (query) {
                props.onNavbarQueryChange({ query, cursorPosition: query.length })
            } else {
                // If we have no component state, then we may have gotten unmounted during a route change.
                const query = props.location.state ? props.location.state.query : ''
                props.onNavbarQueryChange({
                    query,
                    cursorPosition: query.length,
                })
            }
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(authRequired.subscribe(authRequired => this.setState({ authRequired })))
    }

    public componentDidUpdate(prevProps: Props): void {
        if (prevProps.location !== this.props.location) {
            if (!this.props.isSearchRelatedPage) {
                // On a non-search related page or non-repo page, we clear the query in
                // the main query input and interactive mode UI to avoid misleading users
                // that the query is relevant in any way on those pages.
                this.props.onNavbarQueryChange({ query: '', cursorPosition: 0 })
                this.props.onFiltersInQueryChange({})
            }
        }

        if (prevProps.location.search !== this.props.location.search) {
            const query = parseSearchURLQuery(this.props.location.search || '')
            if (query) {
                if (this.props.interactiveSearchMode) {
                    let filtersInQuery: FiltersToTypeAndValue = {}
                    const { filtersInQuery: newFiltersInQuery, navbarQuery } = convertPlainTextToInteractiveQuery(query)
                    filtersInQuery = { ...filtersInQuery, ...newFiltersInQuery }
                    this.props.onNavbarQueryChange({ query: navbarQuery, cursorPosition: navbarQuery.length })

                    this.props.onFiltersInQueryChange(filtersInQuery)
                } else {
                    this.props.onNavbarQueryChange({ query, cursorPosition: query.length })
                }
            }
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let logoSrc = '/.assets/img/sourcegraph-mark.svg'
        let logoLinkClassName = 'global-navbar__logo-link global-navbar__logo-animated'

        const branding = window.context ? window.context.branding : null
        if (branding) {
            if (this.props.isLightTheme) {
                if (branding.light && branding.light.symbol) {
                    logoSrc = branding.light.symbol
                }
            } else if (branding.dark && branding.dark.symbol) {
                logoSrc = branding.dark.symbol
            }
            if (branding.disableSymbolSpin) {
                logoLinkClassName = 'global-navbar__logo-link'
            }
        }

        const logo = <img className="global-navbar__logo" src={logoSrc} />
        const logoLink = !this.state.authRequired ? (
            <Link to="/search" className={logoLinkClassName}>
                {logo}
            </Link>
        ) : (
            <div className={logoLinkClassName}>{logo}</div>
        )
        const navLinks = !this.state.authRequired && !this.props.hideNavLinks && (
            <NavLinks {...this.props} showDotComMarketing={showDotComMarketing} />
        )

        return (
            <div className={`global-navbar ${this.props.lowProfile ? '' : 'global-navbar--bg border-bottom'} py-1`}>
                {this.props.lowProfile ? (
                    <>
                        <div className="flex-1" />
                        {!this.state.authRequired && !this.props.hideNavLinks && (
                            <NavLinks {...this.props} showDotComMarketing={showDotComMarketing} />
                        )}
                    </>
                ) : this.props.hideGlobalSearchInput ? (
                    <>
                        {logoLink}
                        <div className="nav-item flex-1">
                            <Link to="/search" className="nav-link">
                                Search
                            </Link>
                        </div>
                        {navLinks}
                    </>
                ) : (
                    <>
                        {this.props.splitSearchModes && this.props.interactiveSearchMode ? (
                            !this.state.authRequired && (
                                <InteractiveModeInput
                                    {...this.props}
                                    authRequired={this.state.authRequired}
                                    navbarSearchState={this.props.navbarSearchQueryState}
                                    onNavbarQueryChange={this.props.onNavbarQueryChange}
                                    lowProfile={!this.props.isSearchRelatedPage}
                                />
                            )
                        ) : (
                            <>
                                {logoLink}
                                {!this.state.authRequired && (
                                    <div className="global-navbar__search-box-container d-none d-sm-flex flex-row">
                                        {this.props.splitSearchModes && (
                                            <SearchModeToggle
                                                {...this.props}
                                                interactiveSearchMode={this.props.interactiveSearchMode}
                                            />
                                        )}
                                        <SearchNavbarItem
                                            {...this.props}
                                            navbarSearchState={this.props.navbarSearchQueryState}
                                            onChange={this.props.onNavbarQueryChange}
                                            smartSearchField={this.props.smartSearchField}
                                        />
                                    </div>
                                )}
                                {navLinks}
                            </>
                        )}
                    </>
                )}
            </div>
        )
    }
}
