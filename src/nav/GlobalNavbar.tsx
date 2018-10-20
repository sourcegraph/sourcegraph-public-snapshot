import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { authRequired } from '../auth'
import * as GQL from '../backend/graphqlschema'
import {
    ConfigurationCascadeProps,
    ExtensionsControllerProps,
    ExtensionsProps,
} from '../extensions/ExtensionsClientCommonContext'
import { KeybindingsProps } from '../keybindings'
import { parseSearchURLQuery, SearchOptions } from '../search'
import { SearchNavbarItem } from '../search/input/SearchNavbarItem'
import { NavLinks } from './NavLinks'

interface Props extends ConfigurationCascadeProps, ExtensionsProps, ExtensionsControllerProps, KeybindingsProps {
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
        const options = parseSearchURLQuery(props.location.search || '')
        if (options) {
            props.onNavbarQueryChange(options.query)
        } else {
            // If we have no component state, then we may have gotten unmounted during a route change.
            const state: SearchOptions | undefined = props.location.state
            props.onNavbarQueryChange(state ? state.query : '')
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(authRequired.subscribe(authRequired => this.setState({ authRequired })))
    }

    public componentDidUpdate(prevProps: Props): void {
        if (prevProps.location.search !== this.props.location.search) {
            const options = parseSearchURLQuery(this.props.location.search || '')
            if (options) {
                this.props.onNavbarQueryChange(options.query)
            }
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const logo = <img className="global-navbar__logo" src="/.assets/img/sourcegraph-mark.svg" />
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
                        {!this.state.authRequired && (
                            <div className="global-navbar__search-box-container">
                                <SearchNavbarItem
                                    {...this.props}
                                    navbarSearchQuery={this.props.navbarSearchQuery}
                                    onChange={this.props.onNavbarQueryChange}
                                />
                            </div>
                        )}
                    </>
                )}
                {!this.state.authRequired && <NavLinks {...this.props} />}
            </div>
        )
    }
}
