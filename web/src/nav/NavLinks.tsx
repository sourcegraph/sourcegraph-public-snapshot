import * as H from 'history'
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
import { WebActionsNavItems, WebCommandListPopoverButton } from '../components/shared'
import { isDiscussionsEnabled } from '../discussions'
import { ThemeProps } from '../../../shared/src/theme'
import { EventLoggerProps } from '../tracking/eventLogger'
import { fetchAllStatusMessages, StatusMessagesNavItem } from './StatusMessagesNavItem'
import { UserNavItem } from './UserNavItem'
import { CampaignsNavItem } from '../enterprise/campaigns/global/nav/CampaignsNavItem'
import { ThemePreferenceProps } from '../theme'
import {
    KeyboardShortcutsProps,
    KEYBOARD_SHORTCUT_SHOW_COMMAND_PALETTE,
    KEYBOARD_SHORTCUT_SWITCH_THEME,
} from '../keyboardShortcuts/keyboardShortcuts'

interface Props
    extends SettingsCascadeProps,
        KeyboardShortcutsProps,
        ExtensionsControllerProps<'executeCommand' | 'services'>,
        PlatformContextProps<'forceUpdateTooltip' | 'settings'>,
        ThemeProps,
        ThemePreferenceProps,
        EventLoggerProps,
        ActivationProps {
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    showDotComMarketing: boolean
    showCampaigns: boolean
    isSourcegraphDotCom: boolean
}

export class NavLinks extends React.PureComponent<Props> {
    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <ul className="nav-links nav align-items-center pl-2 pr-1">
                {/* Show "Search" link on small screens when GlobalNavbar hides the SearchNavbarItem. */}
                {that.props.location.pathname !== '/search' && (
                    <li className="nav-item d-sm-none">
                        <Link className="nav-link" to="/search">
                            Search
                        </Link>
                    </li>
                )}
                <WebActionsNavItems {...that.props} menu={ContributableMenu.GlobalNav} />
                {that.props.activation && (
                    <li className="nav-item">
                        <ActivationDropdown activation={that.props.activation} history={that.props.history} />
                    </li>
                )}
                {(!that.props.showDotComMarketing || !!that.props.authenticatedUser) && (
                    <li className="nav-item">
                        <Link to="/explore" className="nav-link">
                            Explore
                        </Link>
                    </li>
                )}
                {that.props.showCampaigns && (
                    <li className="nav-item">
                        <CampaignsNavItem />
                    </li>
                )}
                {!that.props.authenticatedUser && (
                    <>
                        <li className="nav-item">
                            <Link to="/extensions" className="nav-link">
                                Extensions
                            </Link>
                        </li>
                        {that.props.location.pathname !== '/sign-in' && (
                            <li className="nav-item mx-1">
                                <Link className="nav-link btn btn-primary" to="/sign-in">
                                    Sign in
                                </Link>
                            </li>
                        )}
                        {that.props.showDotComMarketing && (
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
                {!that.props.isSourcegraphDotCom &&
                    that.props.authenticatedUser &&
                    that.props.authenticatedUser.siteAdmin && (
                        <li className="nav-item">
                            <StatusMessagesNavItem
                                fetchMessages={fetchAllStatusMessages}
                                isSiteAdmin={that.props.authenticatedUser.siteAdmin}
                            />
                        </li>
                    )}
                <li className="nav-item">
                    <WebCommandListPopoverButton
                        {...that.props}
                        buttonClassName="nav-link btn btn-link"
                        menu={ContributableMenu.CommandPalette}
                        keyboardShortcutForShow={KEYBOARD_SHORTCUT_SHOW_COMMAND_PALETTE}
                    />
                </li>
                {that.props.authenticatedUser && (
                    <li className="nav-item">
                        <UserNavItem
                            {...that.props}
                            authenticatedUser={that.props.authenticatedUser}
                            showDotComMarketing={that.props.showDotComMarketing}
                            showDiscussions={isDiscussionsEnabled(that.props.settingsCascade)}
                            keyboardShortcutForSwitchTheme={KEYBOARD_SHORTCUT_SWITCH_THEME}
                        />
                    </li>
                )}
            </ul>
        )
    }
}
