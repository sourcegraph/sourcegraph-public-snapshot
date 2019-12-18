import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { authRequired } from '../auth'
import { parseSearchURLQuery, PatternTypeProps, InteractiveSearchProps } from '../search'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { EventLoggerProps } from '../tracking/eventLogger'
import { showDotComMarketing } from '../util/features'
import { NavLinks } from './NavLinks'
import { ThemeProps } from '../../../shared/src/theme'
import { ThemePreferenceProps } from '../theme'
import { KeyboardShortcutsProps } from '../keyboardShortcuts/keyboardShortcuts'
import { QueryState } from '../search/helpers'
import InteractiveModeInput from '../search/input/interactive/InteractiveModeInput'
import { FiltersToTypeAndValue } from '../../../shared/src/search/interactive/util'
import { SearchModeToggle } from '../search/input/interactive/SearchModeToggle'

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
        InteractiveSearchProps {
    history: H.History
    location: H.Location<{ query: string }>
    authenticatedUser: GQL.IUser | null
    navbarSearchQueryState: QueryState
    onNavbarQueryChange: (queryState: QueryState) => void
    isSourcegraphDotCom: boolean
    showCampaigns: boolean

    /**
     * Whether to use the low-profile form of the navbar, which has no border or background. Used on the search
     * homepage.
     */
    lowProfile: boolean

    filtersInQuery: FiltersToTypeAndValue
    splitSearchModes: boolean
    interactiveSearchMode: boolean
    toggleSearchMode: (event: React.MouseEvent<HTMLAnchorElement>) => void
}

interface State {
    authRequired?: boolean
}

export class GlobalNavbar extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        // Reads initial state from the props (i.e. URL parameters).
        const navbarQuery = parseSearchURLQuery(props.location.search || '', this.props.interactiveSearchMode, true)
        if (navbarQuery) {
            props.onNavbarQueryChange({ query: navbarQuery, cursorPosition: navbarQuery.length })
        } else {
            // If we have no component state, then we may have gotten unmounted during a route change.
            const query = props.location.state ? props.location.state.query : ''
            props.onNavbarQueryChange({
                query,
                cursorPosition: query.length,
            })
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(authRequired.subscribe(authRequired => this.setState({ authRequired })))
    }

    public componentDidUpdate(prevProps: Props): void {
        if (prevProps.location.search !== this.props.location.search) {
            const navbarQuery = parseSearchURLQuery(this.props.location.search || '', false)
            if (navbarQuery) {
                this.props.onNavbarQueryChange({ query: navbarQuery, cursorPosition: navbarQuery.length })
            }
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let logoSrc = '/.assets/img/sourcegraph-mark.svg'
        let logoLinkClassName = 'global-navbar__logo-link global-navbar__logo-animated'

        const { branding } = window.context
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

        return (
            <div className={`global-navbar ${this.props.lowProfile ? '' : 'global-navbar--bg border-bottom'} py-1`}>
                {this.props.lowProfile ? (
                    <>
                        <div className="flex-1" />
                        {!this.state.authRequired && (
                            <NavLinks {...this.props} showDotComMarketing={showDotComMarketing} />
                        )}
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
                                />
                            )
                        ) : (
                            <>
                                {this.state.authRequired ? (
                                    <div className={logoLinkClassName}>{logo}</div>
                                ) : (
                                    <Link to="/search" className={logoLinkClassName}>
                                        {logo}
                                    </Link>
                                )}
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
                                        />
                                    </div>
                                )}
                                {!this.state.authRequired && (
                                    <NavLinks {...this.props} showDotComMarketing={showDotComMarketing} />
                                )}
                            </>
                        )}
                    </>
                )}
            </div>
        )
    }
}
