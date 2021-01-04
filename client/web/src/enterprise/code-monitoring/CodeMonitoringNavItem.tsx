import React from 'react'
import { LinkWithIconOnlyTooltip } from '../../components/LinkWithIconOnlyTooltip'
import { CodeMonitoringLogo } from './CodeMonitoringLogo'

export const CodeMonitoringNavItem: React.FunctionComponent = () => (
    <LinkWithIconOnlyTooltip
        to="/code-monitoring"
        text="Code Monitoring"
        icon={CodeMonitoringLogo}
        tooltip="Code monitoring"
        className="nav-link btn btn-link px-1 text-decoration-none"
        activeClassName="active"
    />
)
