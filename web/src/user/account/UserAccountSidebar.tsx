import AddIcon from '@sourcegraph/icons/lib/Add'
import CityIcon from '@sourcegraph/icons/lib/City'
import FeedIcon from '@sourcegraph/icons/lib/Feed'
import MoonIcon from '@sourcegraph/icons/lib/Moon'
import SignOutIcon from '@sourcegraph/icons/lib/SignOut'
import SunIcon from '@sourcegraph/icons/lib/Sun'
import UserIcon from '@sourcegraph/icons/lib/User'
import * as React from 'react'
import { NavLink, RouteComponentProps } from 'react-router-dom'
import { OrgAvatar } from '../../org/OrgAvatar'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { authExp } from '../../site-admin/SiteAdminAuthenticationProvidersPage'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAreaPageProps } from '../area/UserArea'

interface Props extends UserAreaPageProps, RouteComponentProps<{}> {
    className: string
    isLightTheme: boolean
    externalAuthEnabled: boolean
    onThemeChange: () => void
}

/**
 * Sidebar for user settings pages
 */
export const UserAccountSidebar: React.SFC<Props> = props => {
    if (!props.authenticatedUser) {
        return null
    }

    // When the site admin is viewing another user's settings.
    const siteAdminViewingOtherUser = props.user.id !== props.authenticatedUser.id

    return (
        <div className={`sidebar user-settings-sidebar ${props.className}`}>
            {/* Indicate when the site admin is viewing another user's settings */}
            {siteAdminViewingOtherUser && (
                <SiteAdminAlert className="sidebar__alert">
                    Viewing settings for <strong>{props.user.username}</strong>
                </SiteAdminAlert>
            )}

            <ul className="sidebar__items">
                <li className="sidebar__header">
                    <div className="sidebar__header-icon">
                        <UserIcon className="icon-inline" />
                    </div>
                    <h5 className="sidebar__header-title">User account</h5>
                </li>
                <li className="sidebar__item">
                    <NavLink
                        to={`${props.match.path}/profile`}
                        exact={true}
                        className="sidebar__item-link"
                        activeClassName="sidebar__item--active"
                    >
                        Profile
                    </NavLink>
                </li>
                {!siteAdminViewingOtherUser &&
                    !props.externalAuthEnabled && (
                        <li className="sidebar__item">
                            <NavLink
                                to={`${props.match.path}/account`}
                                exact={true}
                                className="sidebar__item-link"
                                activeClassName="sidebar__item--active"
                            >
                                Password
                            </NavLink>
                        </li>
                    )}
                <li className="sidebar__item">
                    <NavLink
                        to={`${props.match.path}/emails`}
                        exact={true}
                        className="sidebar__item-link"
                        activeClassName="sidebar__item--active"
                    >
                        Emails
                    </NavLink>
                </li>
                {authExp && (
                    <li className="sidebar__item">
                        <NavLink
                            to={`${props.match.path}/external-accounts`}
                            exact={true}
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                        >
                            External accounts
                        </NavLink>
                    </li>
                )}
                {window.context.accessTokensEnabled && (
                    <li className="sidebar__item">
                        <NavLink
                            to={`${props.match.path}/tokens`}
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                        >
                            Access tokens
                        </NavLink>
                    </li>
                )}
            </ul>

            {!siteAdminViewingOtherUser && (
                <div className="user-settings-sidebar__theme-switcher">
                    <a className="sidebar__link" onClick={props.onThemeChange} title="Switch to light theme">
                        <div
                            className={
                                'user-settings-sidebar__theme-switcher--button' +
                                (props.isLightTheme ? ' user-settings-sidebar__theme-switcher--button--selected' : '')
                            }
                        >
                            <SunIcon className="user-settings-sidebar__theme-switcher--icon icon-inline" />
                            Light
                        </div>
                    </a>
                    <a className="sidebar__link" onClick={props.onThemeChange} title="Switch to dark theme">
                        <div
                            className={
                                'user-settings-sidebar__theme-switcher--button' +
                                (!props.isLightTheme ? ' user-settings-sidebar__theme-switcher--button--selected' : '')
                            }
                        >
                            <MoonIcon className="user-settings-sidebar__theme-switcher--icon icon-inline" />
                            Dark
                        </div>
                    </a>
                </div>
            )}

            {(props.user.organizations.nodes.length > 0 || !siteAdminViewingOtherUser) && (
                <>
                    <ul className="sidebar__items">
                        <li className="sidebar__header">
                            <div className="sidebar__header-icon">
                                <CityIcon className="icon-inline" />
                            </div>
                            <h5 className="sidebar__header-title">Organizations</h5>
                        </li>
                        {props.user.organizations.nodes.map(org => (
                            <li className="sidebar__item" key={org.id}>
                                <NavLink
                                    to={`/organizations/${org.name}/settings`}
                                    className="sidebar__item-link"
                                    activeClassName="sidebar__item--active"
                                >
                                    <div className="sidebar__item-icon">
                                        <OrgAvatar org={org.name} />
                                    </div>
                                    <span className="sidebar__item-link-text">{org.name}</span>
                                </NavLink>
                            </li>
                        ))}
                        {!siteAdminViewingOtherUser && (
                            <li className="sidebar__item sidebar__action sidebar__item-action">
                                <NavLink
                                    to="/organizations/new"
                                    className="sidebar__action-button btn"
                                    activeClassName="sidebar__item--active"
                                >
                                    <AddIcon className="icon-inline sidebar__action-icon" />New organization
                                </NavLink>
                            </li>
                        )}
                    </ul>
                    <div className="sidebar__spacer" />
                </>
            )}
            {!siteAdminViewingOtherUser && (
                <div className="sidebar__item sidebar__action">
                    <NavLink
                        to="/api/console"
                        className="sidebar__action-button btn"
                        activeClassName="sidebar__item--active"
                    >
                        <FeedIcon className="icon-inline sidebar__action-icon" /> API console
                    </NavLink>
                </div>
            )}
            {!siteAdminViewingOtherUser && (
                <div className="sidebar__item sidebar__action">
                    <NavLink
                        to={`${props.match.path}/integrations`}
                        exact={true}
                        className="sidebar__action-button btn"
                    >
                        Integrations
                    </NavLink>
                </div>
            )}
            {props.authenticatedUser.siteAdmin && (
                <div className="sidebar__item sidebar__action">
                    <NavLink
                        to="/site-admin"
                        className="sidebar__action-button btn"
                        activeClassName="sidebar__item--active"
                    >
                        Site admin
                    </NavLink>
                </div>
            )}
            {!siteAdminViewingOtherUser &&
                props.authenticatedUser.session &&
                props.authenticatedUser.session.canSignOut && (
                    <div className="sidebar__item sidebar__action">
                        <a href="/-/sign-out" className="sidebar__action-button btn" onClick={logTelemetryOnSignOut}>
                            <SignOutIcon className="icon-inline sidebar__item-action-icon" /> Sign out
                        </a>
                    </div>
                )}
        </div>
    )
}

function logTelemetryOnSignOut(): void {
    eventLogger.log('SignOutClicked')
}
