import * as React from 'react'

import classNames from 'classnames'
import { NavLink } from 'react-router-dom'

export const RepositoryBranchesNavbar: React.FunctionComponent<
    React.PropsWithChildren<{ repo: string; className: string }>
> = ({ repo, className }) => (
    <ul className={classNames('nav', className)}>
        <li className="nav-item">
            <NavLink className="nav-link" exact={true} activeClassName="font-weight-bold" to={`/${repo}/-/branches`}>
                Overview
            </NavLink>
        </li>
        <li className="nav-item">
            <NavLink className="nav-link" activeClassName="font-weight-bold" to={`/${repo}/-/branches/all`}>
                All branches
            </NavLink>
        </li>
    </ul>
)
