import * as React from 'react'

import classNames from 'classnames'
import { NavLink } from 'react-router-dom'

export const RepositoryStatsNavbar: React.FunctionComponent<
    React.PropsWithChildren<{ repo: string; className: string }>
> = ({ repo, className }) => (
    <ul className={classNames('nav', className)}>
        <li className="nav-item">
            <NavLink
                className="nav-link"
                exact={true}
                activeClassName="font-weight-bold"
                to={`/${repo}/-/stats/contributors`}
            >
                Contributors
            </NavLink>
        </li>
    </ul>
)
