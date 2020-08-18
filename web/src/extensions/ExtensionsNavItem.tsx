import React from 'react'
import { LinkWithIconOnlyTooltip } from '../components/LinkWithIconOnlyTooltip'
import { ExtensionsNavIcon } from './icons'

export const ExtensionsNavItem: React.FunctionComponent = () => (
    <LinkWithIconOnlyTooltip
        to="/extensions"
        text="Extensions"
        icon={ExtensionsNavIcon}
        className="nav-link btn btn-link px-1 text-decoration-none"
        activeClassName="active"
    />
)
