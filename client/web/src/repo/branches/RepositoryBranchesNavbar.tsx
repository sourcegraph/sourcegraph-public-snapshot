import type { FC } from 'react'

import classNames from 'classnames'
import { NavLink } from 'react-router-dom'

export const RepositoryBranchesNavbar: FC<{ repo: string; className: string }> = ({ repo, className }) => (
    <ul className={classNames('nav', className)}>
        <li className="nav-item">
            <NavLink
                className={({ isActive }) => classNames('nav-link', isActive && 'font-weight-bold')}
                to={`/${repo}/-/branches`}
                end={true}
            >
                Overview
            </NavLink>
        </li>
        <li className="nav-item">
            <NavLink
                className={({ isActive }) => classNames('nav-link', isActive && 'font-weight-bold')}
                to={`/${repo}/-/branches/all`}
            >
                All branches
            </NavLink>
        </li>
    </ul>
)
