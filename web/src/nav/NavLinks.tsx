import * as H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { ActionsNavItems } from '../../../shared/src/actions/ActionsNavItems'
import { ContributableMenu } from '../../../shared/src/api/protocol'
import { CommandListPopoverButton } from '../../../shared/src/commandPalette/CommandList'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { ActivationDropdown } from '../../../shared/src/components/activation/ActivationDropdown'
import { Link } from '../../../shared/src/components/Link'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { isDiscussionsEnabled } from '../discussions'
import { KeybindingsProps } from '../keybindings'
import { ThemePreferenceProps, ThemeProps } from '../theme'
import { UserNavItem } from './UserNavItem'

interface Props
    extends SettingsCascadeProps,
        KeybindingsProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip'>,
        ThemeProps,
        ThemePreferenceProps,
        ActivationProps {
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    showDotComMarketing: boolean
}

export class NavLinks extends React.PureComponent<Props> {
    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <ul className="nav-links nav align-items-center pl-2 pr-1">
                {/* Show "Search" link on small screens when GlobalNavbar hides the SearchNavbarItem. */}
                {this.props.location.pathname !== '/search' && this.props.location.pathname !== '/welcome' && (
                    <li className="nav-item d-sm-none">
                        <Link className="nav-link" to="/search">
                            Search
                        </Link>
                    </li>
                )}
                {this.props.showDotComMarketing && this.props.location.pathname !== '/welcome' && (
                    <li className="nav-item">
                        <Link to="/welcome" className="nav-link">
                            Welcome
                        </Link>
                    </li>
                )}
                {this.props.showDotComMarketing && this.props.location.pathname === '/welcome' && (
                    <li className="nav-item">
                        <a href="https://docs.sourcegraph.com" className="nav-link" target="_blank">
                            Docs
                        </a>
                    </li>
                )}
                <ActionsNavItems {...this.props} menu={ContributableMenu.GlobalNav} actionItemClass="nav-link" />
                {this.props.activation && (
                    <li className="nav-item">
                        <ActivationDropdown activation={this.props.activation} history={this.props.history} />
                    </li>
                )}
                {(!this.props.showDotComMarketing ||
                    !!this.props.authenticatedUser ||
                    this.props.location.pathname !== '/welcome') && (
                    <li className="nav-item">
                        <Link to="/explore" className="nav-link">
                            Explore
                        </Link>
                    </li>
                )}
                {!this.props.authenticatedUser && (
                    <>
                        {this.props.location.pathname !== '/welcome' && (
                            <li className="nav-item">
                                <Link to="/extensions" className="nav-link">
                                    Extensions
                                </Link>
                            </li>
                        )}
                        {this.props.location.pathname !== '/sign-in' && (
                            <li className="nav-item mx-1">
                                <Link className="nav-link btn btn-primary" to="/sign-in">
                                    Sign in
                                </Link>
                            </li>
                        )}
                        {this.props.showDotComMarketing && (
                            <li className="nav-item">
                                <a href="https://about.sourcegraph.com" className="nav-link">
                                    About
                                </a>
                            </li>
                        )}
                        {this.props.location.pathname !== '/welcome' && (
                            <li className="nav-item">
                                <Link to="/help" className="nav-link">
                                    Help
                                </Link>
                            </li>
                        )}
                    </>
                )}
                {this.props.location.pathname !== '/welcome' && (
                    <CommandListPopoverButton
                        {...this.props}
                        menu={ContributableMenu.CommandPalette}
                        toggleVisibilityKeybinding={this.props.keybindings.commandPalette}
                    />
                )}
                {this.props.authenticatedUser && (
                    <li className="nav-item">
                        <UserNavItem
                            {...this.props}
                            authenticatedUser={this.props.authenticatedUser}
                            showAbout={this.props.showDotComMarketing}
                            showDiscussions={isDiscussionsEnabled(this.props.settingsCascade)}
                        />
                    </li>
                )}
            </ul>
        )
    }
}
