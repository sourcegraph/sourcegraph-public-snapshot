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
import { KeybindingsProps } from '../keybindings'
import { parseSearchURLQuery } from '../search'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { ThemePreferenceProps, ThemeProps } from '../theme'
import { EventLoggerProps } from '../tracking/eventLogger'
import { showDotComMarketing } from '../util/features'
import { NavLinks } from './NavLinks'
interface Props
    extends SettingsCascadeProps,
        PlatformContextProps,
        ExtensionsControllerProps,
        KeybindingsProps,
        EventLoggerProps,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps {
    history: H.History
    location: H.Location
    authenticatedUser: GQL.IUser | null
    navbarSearchQuery: string
    onNavbarQueryChange: (query: string) => void
    isSourcegraphDotCom: boolean

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
            // TODO!(sqs): hacky, prevent from updating on other pages with ?q param
            if (query && !/^\/(threads|checks|codemods|p\/)/.test(this.props.location.pathname)) {
                this.props.onNavbarQueryChange(query)
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
                {!this.props.lowProfile && (
                    <>
                        {this.state.authRequired ? (
                            <div className={logoLinkClassName}>{logo}</div>
                        ) : (
                            <Link to="/search" className={logoLinkClassName}>
                                {logo}
                            </Link>
                        )}
                        {!this.state.authRequired && (
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
                <div className="flex-1" />
                {!this.state.authRequired && (
                    <NavLinks
                        {...this.props}
                        showStatusIndicator={!!window.context.showStatusIndicator}
                        showDotComMarketing={showDotComMarketing}
                    />
                )}
            </div>
        )
    }
}
