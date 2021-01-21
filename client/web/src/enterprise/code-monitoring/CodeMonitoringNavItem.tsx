import React from 'react'
import { LinkWithIconOnlyTooltip } from '../../components/LinkWithIconOnlyTooltip'
import { CodeMonitoringLogo } from './CodeMonitoringLogo'

export const CodeMonitoringNavItem: React.FunctionComponent = () => (
    <LinkWithIconOnlyTooltip
        to="/code-monitoring"
        text="Monitoring Code to see the future"
        icon={CodeMonitoringLogo}
        className="nav-link btn btn-link px-1 text-decoration-none"
        activeClassName="active"
    />
)
