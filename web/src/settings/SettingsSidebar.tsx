import AddIcon from '@sourcegraph/icons/lib/Add'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import GlobeIcon from '@sourcegraph/icons/lib/Globe'
import KeyIcon from '@sourcegraph/icons/lib/Key'
import MoonIcon from '@sourcegraph/icons/lib/Moon'
import SignOutIcon from '@sourcegraph/icons/lib/SignOut'
import SunIcon from '@sourcegraph/icons/lib/Sun'
import UserIcon from '@sourcegraph/icons/lib/User'
import * as H from 'history'
import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { OrgAvatar } from '../org/OrgAvatar'
import { hasTagRecursive } from '../settings/tags'
import { colorTheme, getColorTheme, setColorTheme } from '../settings/theme'
import { eventLogger } from '../tracking/eventLogger'

interface Props {
    history: H.History
    location: H.Location
    user: GQL.IUser | null
}

interface State {
    editorBeta: boolean
    currentUser?: GQL.IUser
    orgs?: GQL.IOrg[]
    isLightTheme: boolean
}

/**
 * Sidebar for settings pages
 */
export class SettingsSidebar extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = { editorBeta: false, isLightTheme: getColorTheme() === 'light' }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(user => {
                // If not logged in, redirect
                if (!user) {
                    // TODO currently we can't redirect here because the initial value will always be `null`
                    // this.props.history.push('/sign-in')
                    return
                }
                const editorBeta = hasTagRecursive(user, 'editor-beta')
                this.setState({ orgs: user.orgs, currentUser: user, editorBeta })
            })
        )

        this.subscriptions.add(colorTheme.subscribe(theme => this.setState({ isLightTheme: theme === 'light' })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="sidebar settings-sidebar">
                <ul className="sidebar__items">
                    <li className="sidebar__header">
                        <div className="sidebar__header-icon">
                            <UserIcon className="icon-inline" />
                        </div>
                        <h5 className="sidebar__header-title">Personal settings</h5>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/settings/profile"
                            exact={true}
                            className={`sidebar__item-link`}
                            activeClassName="sidebar__item--active"
                        >
                            Profile
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/settings/configuration"
                            exact={true}
                            className={`sidebar__item-link`}
                            activeClassName="sidebar__item--active"
                        >
                            Configuration
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/settings/account"
                            exact={true}
                            className={`sidebar__item-link`}
                            activeClassName="sidebar__item--active"
                        >
                            Password
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/settings/emails"
                            exact={true}
                            className={`sidebar__item-link`}
                            activeClassName="sidebar__item--active"
                        >
                            Emails
                        </NavLink>
                    </li>
                    {this.state.editorBeta && (
                        <li className="sidebar__item">
                            <NavLink
                                to="/settings/editor-auth"
                                className="sidebar__item-link"
                                activeClassName="sidebar__item--active"
                            >
                                <KeyIcon className="icon-inline sidebar__item-icon" />Editor authentication
                            </NavLink>
                        </li>
                    )}
                    <li className="sidebar__item">
                        <div className="settings-sidebar__theme-switcher">
                            <a className="sidebar__link" onClick={this.toggleTheme} title="Switch to light theme">
                                <div
                                    className={
                                        'settings-sidebar__theme-switcher--button' +
                                        (this.state.isLightTheme
                                            ? ' settings-sidebar__theme-switcher--button--selected'
                                            : '')
                                    }
                                >
                                    <SunIcon className="settings-sidebar__theme-switcher--icon icon-inline" />
                                    Light
                                </div>
                            </a>
                            <a className="sidebar__link" onClick={this.toggleTheme} title="Switch to dark theme">
                                <div
                                    className={
                                        'settings-sidebar__theme-switcher--button' +
                                        (!this.state.isLightTheme
                                            ? ' settings-sidebar__theme-switcher--button--selected'
                                            : '')
                                    }
                                >
                                    <MoonIcon className="settings-sidebar__theme-switcher--icon icon-inline" />
                                    Dark
                                </div>
                            </a>
                        </div>
                    </li>
                </ul>

                <div className="sidebar__spacer" />

                <ul className="sidebar__items">
                    <li className="sidebar__header">
                        <h5 className="sidebar__header-title ui-title">Organization settings</h5>
                    </li>
                    {this.state.orgs &&
                        this.state.orgs.map(org => (
                            <li className="sidebar__item" key={org.id}>
                                <NavLink
                                    to={`/organizations/${org.name}/settings`}
                                    className="sidebar__item-link"
                                    activeClassName="sidebar__item--active"
                                >
                                    <div className="sidebar__item-icon">
                                        <OrgAvatar org={org.name} />
                                    </div>
                                    {org.name}
                                </NavLink>
                            </li>
                        ))}
                    <li className="sidebar__item sidebar__action sidebar__item-action">
                        <NavLink
                            to="/organizations/new"
                            className="sidebar__action-button btn"
                            activeClassName="sidebar__item--active"
                        >
                            <AddIcon className="icon-inline sidebar__action-icon" />New organization
                        </NavLink>
                    </li>
                </ul>

                <div className="sidebar__spacer" />

                {this.state.editorBeta && (
                    <div className="sidebar__item sidebar__action">
                        <a
                            className="sidebar__action-button btn"
                            target="_blank"
                            href="https://about.sourcegraph.com/docs/editor/setup/"
                        >
                            Download Editor
                        </a>
                    </div>
                )}
                {this.props.user && (
                    <div className="sidebar__item sidebar__action">
                        <NavLink
                            to="/api/explorer"
                            className="sidebar__action-button btn"
                            activeClassName="sidebar__item--active"
                        >
                            <GlobeIcon className="icon-inline sidebar__action-icon" />
                            GraphQL API explorer
                        </NavLink>
                    </div>
                )}
                {this.props.user &&
                    this.props.user.siteAdmin && (
                        <div className="sidebar__item sidebar__action">
                            <NavLink
                                to="/site-admin"
                                className="sidebar__action-button btn"
                                activeClassName="sidebar__item--active"
                            >
                                <GearIcon className="icon-inline sidebar__action-icon" />
                                Site admin
                            </NavLink>
                        </div>
                    )}
                <div className="sidebar__item sidebar__action">
                    <a href="/-/sign-out" className="sidebar__action-button btn" onClick={this.logTelemetryOnSignOut}>
                        <SignOutIcon className="icon-inline sidebar__action-icon" />
                        Sign out
                    </a>
                </div>
            </div>
        )
    }

    private toggleTheme = () => setColorTheme(getColorTheme() === 'light' ? 'dark' : 'light')

    private logTelemetryOnSignOut = (): void => {
        eventLogger.log('SignOutClicked')
    }
}
