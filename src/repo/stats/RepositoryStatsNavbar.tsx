import * as React from 'react'
import { NavLink } from 'react-router-dom'

export const RepositoryStatsNavbar: React.SFC<{ repo: string; className: string }> = ({ repo, className }) => (
    <ul className={`nav ${className}`}>
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
