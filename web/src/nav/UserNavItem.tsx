import * as H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../shared/src/graphql/schema'
import { eventLogger } from '../tracking/eventLogger'
import { UserAvatar } from '../user/UserAvatar'

interface Props {
    location: H.Location
    authenticatedUser: GQL.IUser
    isLightTheme: boolean
    onThemeChange: () => void
    showAbout: boolean
    showDiscussions: boolean
}

interface State {
    isOpen: boolean
}

/**
 * Displays the user's avatar and/or username in the navbar and exposes a dropdown menu with more options for
 * authenticated viewers.
 */
export class UserNavItem extends React.PureComponent<Props, State> {
    public state: State = { isOpen: false }

    private onWindowClick = (event: MouseEvent) => {
        this.setState({ isOpen: false })
    }

    public componentDidMount(): void {
        window.addEventListener('click', this.onWindowClick)
    }

    public componentWillUnmount(): void {
        window.removeEventListener('click', this.onWindowClick)
    }

    private onDropdownMenuClick: React.MouseEventHandler = event => {
        event.preventDefault()
        event.stopPropagation()
        this.setState(prevState => ({ isOpen: !prevState.isOpen }))
    }

    private onThemeChange = () => {
        eventLogger.log(this.props.isLightTheme ? 'DarkThemeClicked' : 'LightThemeClicked')
        this.setState(prevState => ({ isOpen: !prevState.isOpen }), this.props.onThemeChange)
    }

    public render(): JSX.Element | null {
        return (
            <div className="dropdown">
                <a
                    className="nav-link dropdown-toggle bg-transparent d-flex align-items-center"
                    href=""
                    aria-haspopup="true"
                    onClick={this.onDropdownMenuClick}
                >
                    {this.props.authenticatedUser.avatarURL ? (
                        <UserAvatar user={this.props.authenticatedUser} size={48} className="icon-inline" />
                    ) : (
                        <strong>{this.props.authenticatedUser.username}</strong>
                    )}
                </a>
                <div
                    className="dropdown-menu dropdown-menu-right"
                    /* tslint:disable-next-line: jsx-ban-props */
                    style={{ display: this.state.isOpen ? 'block' : 'none' }}
                >
                    <h6 className="dropdown-header py-1">
                        Signed in as <strong>@{this.props.authenticatedUser.username}</strong>
                    </h6>
                    <div className="dropdown-divider" />
                    <Link to={`${this.props.authenticatedUser.url}/account`} className="dropdown-item">
                        Account
                    </Link>
                    <Link to={`${this.props.authenticatedUser.url}/settings`} className="dropdown-item">
                        Settings
                    </Link>
                    <Link to="/extensions" className="dropdown-item">
                        Extensions
                    </Link>
                    {this.props.showDiscussions && (
                        <Link to="/discussions" className="dropdown-item">
                            Discussions
                        </Link>
                    )}
                    <Link to="/search/searches" className="dropdown-item">
                        Saved searches
                    </Link>
                    <button type="button" className="dropdown-item theme-switcher" onClick={this.onThemeChange}>
                        Use {this.props.isLightTheme ? 'dark' : 'light'} theme
                    </button>
                    {window.context.sourcegraphDotComMode ? (
                        <a href="https://docs.sourcegraph.com" target="_blank" className="dropdown-item">
                            Help
                        </a>
                    ) : (
                        <Link to="/help" className="dropdown-item">
                            Help
                        </Link>
                    )}
                    {this.props.authenticatedUser.siteAdmin && (
                        <>
                            <div className="dropdown-divider" />
                            <Link to="/site-admin" className="dropdown-item">
                                Site admin
                            </Link>
                        </>
                    )}
                    {this.props.authenticatedUser.session && this.props.authenticatedUser.session.canSignOut && (
                        <>
                            <div className="dropdown-divider" />
                            <a href="/-/sign-out" className="dropdown-item">
                                Sign out
                            </a>
                        </>
                    )}
                    {this.props.showAbout && (
                        <>
                            <div className="dropdown-divider" />
                            <a href="https://about.sourcegraph.com" target="_blank" className="dropdown-item">
                                About Sourcegraph
                            </a>
                        </>
                    )}
                </div>
            </div>
        )
    }
}
