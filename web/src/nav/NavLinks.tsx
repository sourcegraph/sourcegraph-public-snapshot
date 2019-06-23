import * as H from 'history'
import BellIcon from 'mdi-react/BellIcon'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { ContributableMenu } from '../../../shared/src/api/protocol'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import { ActivationDropdown } from '../../../shared/src/components/activation/ActivationDropdown'
import { Link } from '../../../shared/src/components/Link'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { LinkWithIconOnlyTooltip } from '../components/LinkWithIconOnlyTooltip'
import { WebActionsNavItems, WebCommandListPopoverButton } from '../components/shared'
import { isDiscussionsEnabled } from '../discussions'
import { ChangesIcon } from '../enterprise/changes/icons'
import { ChangesetIcon } from '../enterprise/changesets/icons'
import { ChecksNavItem } from '../enterprise/checks/global/nav/ChecksNavItem'
import { TasksIcon } from '../enterprise/tasks/icons'
import { ThreadsNavItem } from '../enterprise/threads/global/nav/ThreadsNavItem'
import { KeybindingsProps } from '../keybindings'
import { ThemePreferenceProps, ThemeProps } from '../theme'
import { EventLoggerProps } from '../tracking/eventLogger'
import { fetchAllStatusMessages, StatusMessagesNavItem } from './StatusMessagesNavItem'
import { UserNavItem } from './UserNavItem'

interface Props
    extends SettingsCascadeProps,
        KeybindingsProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip'>,
        ThemeProps,
        ThemePreferenceProps,
        EventLoggerProps,
        ActivationProps {
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    showDotComMarketing: boolean
    isSourcegraphDotCom: boolean
    showStatusIndicator: boolean
    className?: string
}

export class NavLinks extends React.PureComponent<Props> {
    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <ul className={`nav-links nav align-items-center pl-2 pr-1 ${this.props.className || ''}`}>
                {/* Show "Search" link on small screens when GlobalNavbar hides the SearchNavbarItem. */}
                {this.props.location.pathname !== '/search' && (
                    <li className="nav-item d-sm-none">
                        <Link className="nav-link" to="/search">
                            Search
                        </Link>
                    </li>
                )}
                <WebActionsNavItems {...this.props} menu={ContributableMenu.GlobalNav} />
                {this.props.activation && (
                    <li className="nav-item">
                        <ActivationDropdown activation={this.props.activation} history={this.props.history} />
                    </li>
                )}
                {(!this.props.showDotComMarketing || !!this.props.authenticatedUser) && (
                    // TODO!(sqs): only show these on enterprise
                    <>
                        <li className="nav-item">
                            <LinkWithIconOnlyTooltip
                                to="/changesets"
                                text="Changesets"
                                icon={ChangesetIcon}
                                className="nav-link btn btn-link px-3 text-decoration-none"
                            />
                        </li>

                        <li className="nav-item">
                            <LinkWithIconOnlyTooltip
                                to="/tasks"
                                text="Tasks"
                                icon={TasksIcon}
                                className="nav-link btn btn-link px-3 text-decoration-none"
                            />
                        </li>
                        <li className="nav-item">
                            <ChecksNavItem className="px-3" />
                        </li>
                        <li className="nav-item mr-1">
                            <ThreadsNavItem className="px-3" />
                        </li>
                        <li className="nav-item">
                            <Link
                                to="/notifications"
                                data-tooltip="Notifications"
                                className="nav-link btn btn-link px-3 text-decoration-none"
                            >
                                <BellIcon className="icon-inline" />
                            </Link>
                        </li>
                        <li className="nav-item d-none">
                            <LinkWithIconOnlyTooltip
                                to="/changes"
                                text="Changes"
                                icon={ChangesIcon}
                                className="nav-link btn btn-link px-3 text-decoration-none"
                            />
                        </li>
                    </>
                )}
                {!this.props.authenticatedUser && (
                    <>
                        <li className="nav-item">
                            <Link to="/extensions" className="nav-link">
                                Extensions
                            </Link>
                        </li>
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
                        <li className="nav-item">
                            <Link to="/help" className="nav-link">
                                Help
                            </Link>
                        </li>
                    </>
                )}
                {!this.props.isSourcegraphDotCom &&
                    this.props.showStatusIndicator &&
                    this.props.authenticatedUser &&
                    this.props.authenticatedUser.siteAdmin && (
                        <li className="nav-item">
                            <StatusMessagesNavItem
                                fetchMessages={fetchAllStatusMessages}
                                isSiteAdmin={this.props.authenticatedUser.siteAdmin}
                            />
                        </li>
                    )}
                <li className="nav-item">
                    <WebCommandListPopoverButton
                        {...this.props}
                        buttonClassName="nav-link btn btn-link"
                        menu={ContributableMenu.CommandPalette}
                        toggleVisibilityKeybinding={this.props.keybindings.commandPalette}
                    />
                </li>
                {this.props.authenticatedUser && (
                    <li className="nav-item">
                        <UserNavItem
                            {...this.props}
                            authenticatedUser={this.props.authenticatedUser}
                            showDotComMarketing={this.props.showDotComMarketing}
                            showDiscussions={isDiscussionsEnabled(this.props.settingsCascade)}
                            switchThemeKeybinding={this.props.keybindings.switchTheme}
                        />
                    </li>
                )}
            </ul>
        )
    }
}
