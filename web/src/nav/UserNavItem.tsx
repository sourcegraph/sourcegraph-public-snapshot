import * as H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
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

    public componentDidUpdate(prevProps: Props): void {
        // Close dropdown after clicking on a dropdown item.
        if (this.state.isOpen && this.props.location !== prevProps.location) {
            this.setState({ isOpen: false })
        }
    }

    public render(): JSX.Element | null {
        return (
            <ButtonDropdown isOpen={this.state.isOpen} toggle={this.toggleIsOpen} className="nav-link py-0">
                <DropdownToggle
                    caret={true}
                    className="bg-transparent d-flex align-items-center e2e-user-nav-item-toggle"
                    nav={true}
                >
                    {this.props.authenticatedUser.avatarURL ? (
                        <UserAvatar user={this.props.authenticatedUser} size={48} className="icon-inline" />
                    ) : (
                        <strong>{this.props.authenticatedUser.username}</strong>
                    )}
                </DropdownToggle>
                <DropdownMenu right={true}>
                    <DropdownItem header={true} className="py-1">
                        Signed in as <strong>@{this.props.authenticatedUser.username}</strong>
                    </DropdownItem>
                    <DropdownItem divider={true} />
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
                    <button
                        type="button"
                        className="dropdown-item e2e-user-nav-item__theme"
                        onClick={this.onThemeChange}
                    >
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
                            <DropdownItem divider={true} />
                            <Link to="/site-admin" className="dropdown-item">
                                Site admin
                            </Link>
                        </>
                    )}
                    {this.props.authenticatedUser.session && this.props.authenticatedUser.session.canSignOut && (
                        <>
                            <DropdownItem divider={true} />
                            <a href="/-/sign-out" className="dropdown-item">
                                Sign out
                            </a>
                        </>
                    )}
                    {this.props.showAbout && (
                        <>
                            <DropdownItem divider={true} />
                            <a href="https://about.sourcegraph.com" target="_blank" className="dropdown-item">
                                About Sourcegraph
                            </a>
                        </>
                    )}
                </DropdownMenu>
            </ButtonDropdown>
        )
    }

    private toggleIsOpen = () => this.setState(prevState => ({ isOpen: !prevState.isOpen }))

    private onThemeChange = () => {
        eventLogger.log(this.props.isLightTheme ? 'DarkThemeClicked' : 'LightThemeClicked')
        this.setState(prevState => ({ isOpen: !prevState.isOpen }), this.props.onThemeChange)
    }
}
