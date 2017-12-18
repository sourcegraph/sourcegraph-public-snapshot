import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as H from 'history'
import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'

interface Props {
    history: H.History
    location: H.Location
}

interface State {}

/**
 * Sidebar for the site admin area.
 */
export class Sidebar extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-sidebar">
                <div className="site-admin-sidebar__header">
                    <div className="site-admin-sidebar__header-icon">
                        <ServerIcon className="icon-inline" />
                    </div>
                    <h5 className="site-admin-sidebar__header-title">Site Admin</h5>
                </div>
                <ul className="site-admin-sidebar__items">
                    <li className="site-admin-sidebar__item">
                        <NavLink
                            to="/site-admin"
                            className="site-admin-sidebar__item-link"
                            activeClassName="site-admin-sidebar__item--active"
                            exact={true}
                        >
                            Overview
                        </NavLink>
                    </li>
                    <li className="site-admin-sidebar__item">
                        <NavLink
                            to="/site-admin/config"
                            className="site-admin-sidebar__item-link"
                            activeClassName="site-admin-sidebar__item--active"
                            exact={true}
                        >
                            Configuration
                        </NavLink>
                    </li>
                    <li className="site-admin-sidebar__item">
                        <NavLink
                            to="/site-admin/repositories"
                            className="site-admin-sidebar__item-link"
                            activeClassName="site-admin-sidebar__item--active"
                            exact={true}
                        >
                            Repositories
                        </NavLink>
                    </li>
                    <li className="site-admin-sidebar__item">
                        <NavLink
                            to="/site-admin/organizations"
                            className="site-admin-sidebar__item-link"
                            activeClassName="site-admin-sidebar__item--active"
                            exact={true}
                        >
                            Organizations
                        </NavLink>
                    </li>
                    <li className="site-admin-sidebar__item">
                        <NavLink
                            to="/site-admin/users"
                            className="site-admin-sidebar__item-link"
                            activeClassName="site-admin-sidebar__item--active"
                            exact={true}
                        >
                            Users
                        </NavLink>
                    </li>
                    <li className="site-admin-sidebar__item">
                        <NavLink
                            to="/site-admin/analytics"
                            className="site-admin-sidebar__item-link"
                            activeClassName="site-admin-sidebar__item--active"
                            exact={true}
                        >
                            Analytics
                        </NavLink>
                    </li>
                </ul>
            </div>
        )
    }
}
