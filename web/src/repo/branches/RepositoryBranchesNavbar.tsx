import * as React from 'react'
import { NavLink } from 'react-router-dom'

export const RepositoryBranchesNavbar: React.SFC<{ repo: string; className: string }> = ({ repo, className }) => (
    <ul className={`nav nav-pills ${className}`}>
        <li className="nav-item">
            <NavLink className="nav-link" exact={true} activeClassName="active" to={`/${repo}/-/branches`}>
                Overview
            </NavLink>
        </li>
        <li className="nav-item">
            <NavLink className="nav-link" activeClassName="active" to={`/${repo}/-/branches/all`}>
                All branches
            </NavLink>
        </li>
    </ul>
)
