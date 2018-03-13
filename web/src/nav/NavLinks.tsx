import ChevronDownIcon from '@sourcegraph/icons/lib/ChevronDown'
import ChevronUpIcon from '@sourcegraph/icons/lib/ChevronUp'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { ThemeSwitcher } from '../components/ThemeSwitcher'
import { SearchHelp } from '../search/SearchHelp'
import { eventLogger } from '../tracking/eventLogger'
import { UserAvatar } from '../user/UserAvatar'
import { canListAllRepositories, showDotComMarketing } from '../util/features'

interface Props {
    location: H.Location
    isLightTheme: boolean
    onThemeChange: () => void
    showScopes?: boolean
    onShowScopes?: () => void
    className?: string
}

interface State {
    user?: GQL.IUser | null
}

const isGQLUser = (val: any): val is GQL.IUser => val && typeof val === 'object' && val.__typename === 'User'

export class NavLinks extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(user => {
                this.setState({ user })
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private onClickInstall = (): void => {
        eventLogger.log('InstallSourcegraphServerCTAClicked', {
            location_on_page: 'Navbar',
        })
    }

    private onShowScopes: React.MouseEventHandler<HTMLAnchorElement> = e => {
        e.preventDefault()
        if (this.props.onShowScopes) {
            this.props.onShowScopes()
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className={`nav-links ${this.props.className || ''}`}>
                {this.props.onShowScopes && (
                    <a
                        className="nav-links__link nav-links__scopes-toggle"
                        onClick={this.onShowScopes}
                        data-tooltip={this.props.showScopes ? 'Hide scopes' : 'Show scopes'}
                        href=""
                    >
                        {this.props.showScopes ? (
                            <ChevronUpIcon className="icon-inline" />
                        ) : (
                            <ChevronDownIcon className="icon-inline" />
                        )}
                    </a>
                )}
                {showDotComMarketing && (
                    <a
                        href="https://about.sourcegraph.com"
                        className="nav-links__border-link nav-links__ad"
                        onClick={this.onClickInstall}
                        title="Install self-hosted Sourcegraph Server to search your own code"
                    >
                        Install <span className="nav-links__widescreen-only">Sourcegraph Server</span>
                    </a>
                )}
                {this.state.user && (
                    <Link to="/search/searches" className="nav-links__link">
                        Searches
                    </Link>
                )}
                {canListAllRepositories && (
                    <Link to="/explore" className="nav-links__link">
                        Explore
                    </Link>
                )}
                {this.state.user &&
                    this.state.user.siteAdmin && (
                        <Link to="/site-admin" className="nav-links__link">
                            Admin
                        </Link>
                    )}
                {this.state.user && (
                    <Link className="nav-links__link nav-links__link-user" to="/settings/profile">
                        {isGQLUser(this.state.user) && this.state.user.avatarURL ? (
                            <UserAvatar size={64} />
                        ) : isGQLUser(this.state.user) ? (
                            this.state.user.username
                        ) : (
                            'Profile'
                        )}
                    </Link>
                )}
                {!this.state.user &&
                    this.props.location.pathname !== '/sign-in' && (
                        <Link className="nav-links__link btn btn-primary" to="/sign-in">
                            Sign in
                        </Link>
                    )}
                <ThemeSwitcher {...this.props} className="nav-links__theme-switcher" />
                {showDotComMarketing && (
                    <a href="https://about.sourcegraph.com" className="nav-links__link">
                        About
                    </a>
                )}
                <SearchHelp className="nav-links__link" />
            </div>
        )
    }
}
