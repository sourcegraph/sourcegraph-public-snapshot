import React from 'react'

import { NavItem, NavLink } from '../nav'
import type { NavLinkProps } from '../nav/NavBar'

import { BatchChangesIconNav } from './icons'

interface Props extends Pick<NavLinkProps, 'variant'> {
    // Nothing for now.
}

/**
 * An item in {@link GlobalNavbar} that links to the batch changes area.
 */
export const BatchChangesNavItem: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ variant }) => (
    <NavItem icon={BatchChangesIconNav}>
        <NavLink to="/batch-changes" variant={variant}>
            Batch Changes
        </NavLink>
    </NavItem>
)
