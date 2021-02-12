import React from 'react'
import { LinkWithIcon } from '../../components/LinkWithIcon'
import { CodeMonitoringLogo } from './CodeMonitoringLogo'

export const CodeMonitoringNavItem: React.FunctionComponent = () => (
    <LinkWithIcon
        to="/code-monitoring"
        text="Monitoring"
        icon={CodeMonitoringLogo}
        className="nav-link btn btn-link px-1 text-decoration-none"
        activeClassName="active"
    />
)
