import FeedIcon from '@sourcegraph/icons/lib/Feed'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as H from 'history'
import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { Subscription } from 'rxjs'

interface Props {
    history: H.History
    location: H.Location
    className: string
}

interface State {}

/** Feature flag for showing the comments site admin page link in the sidebar. */
const showComments = localStorage.getItem('showComments') !== null

/**
 * Sidebar for the site admin area.
 */
export class SiteAdminSidebar extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className={`sidebar site-admin-sidebar ${this.props.className}`}>
                <ul className="sidebar__items">
                    <li className="sidebar__header">
                        <div className="sidebar__header-icon">
                            <ServerIcon className="icon-inline" />
                        </div>
                        <h5 className="sidebar__header-title">Site Admin</h5>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Overview
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/configuration"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Configuration
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/repositories"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Repositories
                        </NavLink>
                    </li>
                </ul>
                <ul className="sidebar__items">
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/users"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Users
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/organizations"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Organizations
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/global-settings"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Global settings
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/code-intelligence"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Code intelligence
                        </NavLink>
                    </li>
                </ul>
                <ul className="sidebar__items">
                    <li className="sidebar__header">
                        <div className="sidebar__header-icon">
                            <LockIcon className="icon-inline" />
                        </div>
                        <h5 className="sidebar__header-title">Auth</h5>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/auth/providers"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Providers
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/auth/external-accounts"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            External accounts
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/tokens"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Access tokens
                        </NavLink>
                    </li>
                </ul>
                {showComments && (
                    <ul className="sidebar__items">
                        <li className="sidebar__item">
                            <NavLink
                                to="/site-admin/threads"
                                className="sidebar__item-link"
                                activeClassName="sidebar__item--active"
                                exact={true}
                            >
                                Comments
                            </NavLink>
                        </li>
                    </ul>
                )}
                <ul className="sidebar__items">
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/updates"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Updates
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/analytics"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Analytics
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/surveys"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            User surveys
                        </NavLink>
                    </li>
                    <li className="sidebar__item">
                        <NavLink
                            to="/site-admin/pings"
                            className="sidebar__item-link"
                            activeClassName="sidebar__item--active"
                            exact={true}
                        >
                            Pings
                        </NavLink>
                    </li>
                </ul>
                <div className="sidebar__item sidebar__action">
                    <NavLink
                        to="/api/console"
                        className="sidebar__action-button btn"
                        activeClassName="sidebar__item--active"
                    >
                        <FeedIcon className="icon-inline sidebar__action-icon" />
                        API console
                    </NavLink>
                </div>
                <div className="sidebar__item sidebar__action">
                    <a href="/-/debug/" className="sidebar__action-button btn btn-link">
                        Instrumentation
                    </a>
                </div>
            </div>
        )
    }
}
