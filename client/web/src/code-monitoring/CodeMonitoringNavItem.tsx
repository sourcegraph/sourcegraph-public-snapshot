import React from 'react'

import { LinkWithIcon } from '../components/LinkWithIcon'

import { CodeMonitoringLogo } from './CodeMonitoringLogo'

export const CodeMonitoringNavItem: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <LinkWithIcon
        to="/code-monitoring"
        text="Monitoring"
        icon={CodeMonitoringLogo}
        className="nav-link text-decoration-none"
        activeClassName="active"
    />
)
