import GlobeIcon from '@sourcegraph/icons/lib/Globe'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'

interface Props extends RouteComponentProps<any> {
    className: string
    repo?: GQL.IRepository
}

/**
 * Sidebar for repository settings pages.
 */
export const RepoSettingsSidebar: React.SFC<Props> = (props: Props) =>
    props.repo ? (
        <div className={`sidebar repo-sidebar ${props.className}`}>
            <ul className="sidebar__items">
                <div className="sidebar__header">
                    <h5 className="sidebar__header-title">Repository settings</h5>
                </div>
                <li className="sidebar__item">
                    <NavLink
                        to={`/${props.repo.uri}/-/settings`}
                        exact={true}
                        className="sidebar__item-link"
                        activeClassName="sidebar__item--active"
                    >
                        Options
                    </NavLink>
                </li>
                <li className="sidebar__item">
                    <NavLink
                        to={`/${props.repo.uri}/-/settings/index`}
                        exact={true}
                        className="sidebar__item-link"
                        activeClassName="sidebar__item--active"
                    >
                        Indexing
                    </NavLink>
                </li>
                <li className="sidebar__item">
                    <NavLink
                        to={`/${props.repo.uri}/-/settings/mirror`}
                        exact={true}
                        className="sidebar__item-link"
                        activeClassName="sidebar__item--active"
                    >
                        Mirroring
                    </NavLink>
                </li>
            </ul>
            <div className="sidebar__item sidebar__action">
                <Link to="/api/console" className="sidebar__action-button btn">
                    <GlobeIcon className="icon-inline sidebar__action-icon" />
                    API console
                </Link>
            </div>
        </div>
    ) : (
        <div className={`sidebar repo-sidebar ${props.className}`} />
    )
