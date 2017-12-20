import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { ThemeSwitcher } from '../components/ThemeSwitcher'
import { UserAvatar } from '../settings/user/UserAvatar'
import { eventLogger } from '../tracking/eventLogger'
import { canListAllRepositories, showDotComMarketing } from '../util/features'

interface Props {
    location: H.Location
    onToggleTheme: () => void
    isLightTheme: boolean
}

interface State {
    user: GQL.IUser | ImmutableUser | null
}

const isGQLUser = (val: any): val is GQL.IUser => val && typeof val === 'object' && val.__typename === 'User'

export class NavLinks extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            user: window.context.user,
        }
    }

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

    public render(): JSX.Element | null {
        return (
            <div className="nav-links">
                {showDotComMarketing && (
                    <a
                        href="https://about.sourcegraph.com"
                        className="nav-links__border-link"
                        onClick={this.onClickInstall}
                    >
                        Install Sourcegraph Server
                    </a>
                )}
                {this.state.user && (
                    <Link to="/search/queries" className="nav-links__link">
                        Queries
                    </Link>
                )}
                {canListAllRepositories && (
                    <Link to="/browse" className="nav-links__link">
                        Browse
                    </Link>
                )}
                {window.context.user &&
                    window.context.user.IsAdmin && (
                        <Link to="/site-admin" className="nav-links__link">
                            Admin
                        </Link>
                    )}
                {this.state.user && (
                    <Link className="nav-links__link nav-links__link-user" to="/settings">
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
                <ThemeSwitcher onToggleTheme={this.props.onToggleTheme} isLightTheme={this.props.isLightTheme} />
            </div>
        )
    }
}
