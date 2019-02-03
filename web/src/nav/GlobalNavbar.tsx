import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { authRequired } from '../auth'
import { KeybindingsProps } from '../keybindings'
import { parseSearchURLQuery } from '../search'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { showDotComMarketing } from '../util/features'
import { NavLinks } from './NavLinks'

interface Props extends SettingsCascadeProps, PlatformContextProps, ExtensionsControllerProps, KeybindingsProps {
    history: H.History
    location: H.Location
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
    onThemeChange: () => void
    navbarSearchQuery: string
    onNavbarQueryChange: (query: string) => void

    /**
     * Whether to use the low-profile form of the navbar, which has no border or background. Used on the search
     * homepage.
     */
    lowProfile: boolean
}

interface State {
    authRequired?: boolean
}

export class GlobalNavbar extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        /**
         * Reads initial state from the props (i.e. URL parameters).
         */
        const query = parseSearchURLQuery(props.location.search || '')
        if (query) {
            props.onNavbarQueryChange(query)
        } else {
            // If we have no component state, then we may have gotten unmounted during a route change.
            props.onNavbarQueryChange(props.location.state ? props.location.state.query : '')
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(authRequired.subscribe(authRequired => this.setState({ authRequired })))
    }

    public componentDidUpdate(prevProps: Props): void {
        if (prevProps.location.search !== this.props.location.search) {
            const query = parseSearchURLQuery(this.props.location.search || '')
            if (query) {
                this.props.onNavbarQueryChange(query)
            }
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let logoSrc: string
        const showFullLogo = this.props.location.pathname === '/welcome'
        if (showFullLogo) {
            logoSrc = this.props.isLightTheme
                ? '/.assets/img/sourcegraph-light-head-logo.svg'
                : '/.assets/img/sourcegraph-head-logo.svg'
        } else {
            logoSrc = '/.assets/img/sourcegraph-mark.svg'
        }

        const logo = (
            <img className={`global-navbar__logo ${showFullLogo ? 'global-navbar__logo--full' : ''}`} src={logoSrc} />
        )
        return (
            <div className={`global-navbar ${this.props.lowProfile ? '' : 'global-navbar--bg'}`}>
                {this.props.lowProfile ? (
                    <div />
                ) : (
                    <>
                        {this.state.authRequired ? (
                            <div className="global-navbar__logo-link">{logo}</div>
                        ) : (
                            <Link to="/search" className="global-navbar__logo-link">
                                {logo}
                            </Link>
                        )}
                        {!this.state.authRequired && this.props.location.pathname !== '/welcome' && (
                            <div className="global-navbar__search-box-container d-none d-sm-flex">
                                <SearchNavbarItem
                                    {...this.props}
                                    navbarSearchQuery={this.props.navbarSearchQuery}
                                    onChange={this.props.onNavbarQueryChange}
                                />
                            </div>
                        )}
                    </>
                )}
                {!this.state.authRequired && <NavLinks {...this.props} showDotComMarketing={showDotComMarketing} />}
            </div>
        )
    }
}
