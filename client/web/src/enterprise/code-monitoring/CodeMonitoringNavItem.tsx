import React from 'react'
import VideoInputAntennaIcon from 'mdi-react/VideoInputAntennaIcon'
import { LinkWithIconOnlyTooltip } from '../../components/LinkWithIconOnlyTooltip'

export const CodeMonitoringNavItem: React.FunctionComponent = () => (
    <LinkWithIconOnlyTooltip
        to="/code-monitoring"
        text="Code Monitoring"
        icon={VideoInputAntennaIcon}
        className="nav-link btn btn-link px-1 text-decoration-none"
        activeClassName="active"
    />
)
