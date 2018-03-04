import FeedIcon from '@sourcegraph/icons/lib/Feed'
import GlobeIcon from '@sourcegraph/icons/lib/Globe'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'

interface Props extends RouteComponentProps<any> {
    className: string
    repo?: GQL.IRepository
    routePrefix: string
}

/**
 * Sidebar for repository graph pages.
 */
export const RepositoryGraphSidebar: React.SFC<Props> = (props: Props) =>
    props.repo ? (
        <div className={`sidebar repository-graph-sidebar ${props.className}`}>
            <ul className="sidebar__items">
                <div className="sidebar__header">
                    <div className="sidebar__header-icon">
                        <GlobeIcon className="icon-inline" />
                    </div>
                    <h5 className="sidebar__header-title">Repository graph</h5>
                </div>
                <li className="sidebar__item">
                    <NavLink
                        to={`${props.routePrefix}/-/graph`}
                        exact={true}
                        className="sidebar__item-link"
                        activeClassName="sidebar__item--active"
                    >
                        Overview
                    </NavLink>
                </li>
                <li className="sidebar__item">
                    <NavLink
                        to={`${props.routePrefix}/-/graph/packages`}
                        exact={true}
                        className="sidebar__item-link"
                        activeClassName="sidebar__item--active"
                    >
                        Packages
                    </NavLink>
                </li>
                <li className="sidebar__item">
                    <NavLink
                        to={`${props.routePrefix}/-/graph/dependencies`}
                        exact={true}
                        className="sidebar__item-link"
                        activeClassName="sidebar__item--active"
                    >
                        Dependencies
                    </NavLink>
                </li>
            </ul>
            <div className="sidebar__item sidebar__action">
                <Link to="/api/console" className="sidebar__action-button btn">
                    <FeedIcon className="icon-inline sidebar__action-icon" />
                    API console
                </Link>
            </div>
        </div>
    ) : (
        <div className={`sidebar repository-graph-sidebar ${props.className}`} />
    )
