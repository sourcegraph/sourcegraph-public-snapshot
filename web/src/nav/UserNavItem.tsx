import { Shortcut } from '@slimsag/react-shortcuts'
import * as H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import * as GQL from '../../../shared/src/graphql/schema'
import { KeyboardShortcut } from '../../../shared/src/keyboardShortcuts'
import { ThemeProps } from '../../../shared/src/theme'
import { UserAvatar } from '../user/UserAvatar'
import { ThemePreferenceProps, ThemePreference } from '../theme'
import { Menu, MenuButton, MenuList, MenuLink, MenuItem, MenuPopover } from '@reach/menu-button'

interface Props extends ThemeProps, ThemePreferenceProps {
    location: H.Location
    authenticatedUser: Pick<
        GQL.IUser,
        'username' | 'avatarURL' | 'settingsURL' | 'organizations' | 'siteAdmin' | 'session'
    >
    showDotComMarketing: boolean
    keyboardShortcutForSwitchTheme?: KeyboardShortcut
}

interface State {
    isOpen: boolean
}

/**
 * Displays the user's avatar and/or username in the navbar and exposes a dropdown menu with more options for
 * authenticated viewers.
 */
export class UserNavItem extends React.PureComponent<Props, State> {
    private supportsSystemTheme = Boolean(
        window.matchMedia && window.matchMedia('not all and (prefers-color-scheme), (prefers-color-scheme)').matches
    )

    public state: State = { isOpen: false }

    public componentDidUpdate(previousProps: Props): void {
        // Close dropdown after clicking on a dropdown item.
        if (this.state.isOpen && this.props.location !== previousProps.location) {
            /* eslint react/no-did-update-set-state: warn */
            this.setState({ isOpen: false })
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className="user-nav-item">
                <Menu>
                    <MenuButton>
                        {this.props.authenticatedUser.avatarURL ? (
                            <UserAvatar user={this.props.authenticatedUser} size={48} className="icon-inline" />
                        ) : (
                            <strong>{this.props.authenticatedUser.username}</strong>
                        )}
                    </MenuButton>
                    <MenuList tabIndex={0} className="user-nav-item">
                        <MenuItem onSelect={() => {}} tabIndex={-1} className="dropdown-item">
                            Signed in as <strong>@{this.props.authenticatedUser.username}</strong>
                        </MenuItem>
                        <MenuLink
                            onSelect={() => {}}
                            as={Link}
                            to={this.props.authenticatedUser.settingsURL!}
                            tabIndex={0}
                            className="dropdown-item"
                        >
                            Settings
                        </MenuLink>
                        <MenuLink onSelect={() => {}} as={Link} to="/extensions" tabIndex={0} className="dropdown-item">
                            Extensions
                        </MenuLink>
                        <MenuLink
                            onSelect={() => {}}
                            as={Link}
                            to={`/users/${this.props.authenticatedUser.username}/searches`}
                            tabIndex={0}
                            className="dropdown-item"
                        >
                            Saved searches
                        </MenuLink>
                        {/* <MenuItem> */}
                        <div className="px-2 py-1">
                            <div className="d-flex align-items-center">
                                <div className="mr-2">Theme</div>
                                <MenuItem
                                    onSelect={e => {
                                        // e.preventDefault()
                                    }}
                                >
                                    <select
                                        className="custom-select custom-select-sm e2e-theme-toggle"
                                        onChange={this.onThemeChange}
                                        value={this.props.themePreference}
                                    >
                                        <option value={ThemePreference.Light}>Light</option>
                                        <option value={ThemePreference.Dark}>Dark</option>
                                        <option value={ThemePreference.System}>System</option>
                                    </select>
                                </MenuItem>
                            </div>
                            {this.props.themePreference === ThemePreference.System && !this.supportsSystemTheme && (
                                <div className="text-wrap">
                                    <small>
                                        <a
                                            href="https://caniuse.com/#feat=prefers-color-scheme"
                                            className="text-warning"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                        >
                                            Your browser does not support the system theme.
                                        </a>
                                    </small>
                                </div>
                            )}
                            {this.props.keyboardShortcutForSwitchTheme &&
                                this.props.keyboardShortcutForSwitchTheme.keybindings.map((keybinding, i) => (
                                    <Shortcut key={i} {...keybinding} onMatch={this.onThemeCycle} />
                                ))}
                        </div>
                        {this.props.authenticatedUser.organizations.nodes.length > 0 && (
                            <>
                                <hr />
                                <MenuItem onSelect={() => {}}>Organizations</MenuItem>
                                {this.props.authenticatedUser.organizations.nodes.map(org => (
                                    <MenuLink
                                        as={Link}
                                        key={org.id}
                                        to={org.settingsURL || org.url}
                                        className="dropdown-item"
                                    >
                                        {org.displayName || org.name}
                                    </MenuLink>
                                ))}
                            </>
                        )}
                        {this.props.authenticatedUser.siteAdmin && (
                            <MenuLink as={Link} to="/site-admin" className="dropdown-item">
                                Site admin
                            </MenuLink>
                        )}
                        {this.props.showDotComMarketing ? (
                            // eslint-disable-next-line react/jsx-no-target-blank
                            <MenuLink href="https://docs.sourcegraph.com" target="_blank" className="dropdown-item">
                                Help
                            </MenuLink>
                        ) : (
                            <MenuLink as={Link} to="/help" className="dropdown-item">
                                Help
                            </MenuLink>
                        )}
                        {this.props.authenticatedUser.session && this.props.authenticatedUser.session.canSignOut && (
                            <MenuLink href="/-/sign-out" className="dropdown-item">
                                Sign out
                            </MenuLink>
                        )}
                        {this.props.showDotComMarketing && (
                            <>
                                <hr />
                                {/* eslint-disable-next-line react/jsx-no-target-blank */}
                                <MenuLink
                                    href="https://about.sourcegraph.com"
                                    target="_blank"
                                    className="dropdown-item"
                                >
                                    About Sourcegraph
                                </MenuLink>
                            </>
                        )}
                    </MenuList>
                </Menu>
            </div>
        )
    }

    private toggleIsOpen = (): void => this.setState(previousState => ({ isOpen: !previousState.isOpen }))

    private onThemeChange: React.ChangeEventHandler<HTMLSelectElement> = event => {
        this.props.onThemePreferenceChange(event.target.value as ThemePreference)
    }

    private onThemeCycle = (): void => {
        this.props.onThemePreferenceChange(
            this.props.themePreference === ThemePreference.Dark ? ThemePreference.Light : ThemePreference.Dark
        )
    }
}
