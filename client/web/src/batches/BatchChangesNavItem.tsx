import React from 'react'

import { useLocation } from 'react-router-dom'

import { ProductStatusBadge } from '@sourcegraph/wildcard'

import { NavItem, NavLink } from '../nav'

import { BatchChangesIconNav } from './icons'

interface Props {
    // Nothing for now.
}

/**
 * An item in {@link GlobalNavbar} that links to the batch changes area.
 */
export const BatchChangesNavItem: React.FunctionComponent<Props> = () => {
    const { search } = useLocation()
    const searchParameters = new URLSearchParams(search)
    const kind = searchParameters.get('kind')
    const showExperimentalBade = kind?.startsWith('goChecker')

    return (
        <NavItem icon={BatchChangesIconNav}>
            <NavLink to="/batch-changes">
                Batch Changes
                {showExperimentalBade && <ProductStatusBadge className="ml-1" status="experimental" />}
            </NavLink>
        </NavItem>
    )
}
