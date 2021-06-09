import React from 'react'

import { NavItem, NavLink } from '@sourcegraph/wildcard/src/components/NavBar'

import { BatchChangesIconNav } from './icons'

interface Props {
    isSourcegraphDotCom: boolean
}

/**
 * An item in {@link GlobalNavbar} that links to the batch changes area.
 */
export const BatchChangesNavItem: React.FunctionComponent<Props> = ({ isSourcegraphDotCom }) => (
    <NavItem icon={BatchChangesIconNav}>
        <NavLink
            to={isSourcegraphDotCom ? 'https://about.sourcegraph.com/batch-changes' : '/batch-changes'}
            external={isSourcegraphDotCom}
        >
            Batch Changes
        </NavLink>
    </NavItem>
)
