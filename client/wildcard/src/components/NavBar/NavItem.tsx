import React from 'react'
import { NavLink } from 'react-router-dom'

export interface NavItemProps {
    icon?: React.FunctionComponent<unknown>
    children: JSX.Element
    to: string
}

export const NavItem: React.FunctionComponent<NavItemProps> = ({ icon, children, to }): JSX.Element => (
    <NavLink to={to}>
        {icon}
        <span>{children}</span>
    </NavLink>
)
