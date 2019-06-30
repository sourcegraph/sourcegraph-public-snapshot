import BellIcon from 'mdi-react/BellIcon'
import EyeIcon from 'mdi-react/EyeIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { Timeline } from '../../../../components/timeline/Timeline'
import { GitPullRequestIcon } from '../../../../util/octicons'
import { StatusAreaContext } from '../StatusArea'
import { StatusNotificationSettingsDropdownButton } from './stateBar/StatusNotificationSettingsDropdownButton'
import { StatusStateBar } from './stateBar/StatusStateBar'
import { StatusBreadcrumbs } from './StatusBreadcrumbs'

interface Props extends Pick<StatusAreaContext, 'status' | 'statusURL' | 'statusesURL'> {
    className?: string
}

/**
 * An overview of a status.
 */
export const StatusOverview: React.FunctionComponent<Props> = ({ status, statusURL, statusesURL, className = '' }) => (
    <div className={`status-overview ${className || ''}`}>
        <StatusBreadcrumbs status={status} statusURL={statusURL} statusesURL={statusesURL} className="py-3" />
        <hr className="my-0" />
        <h2 className="my-3 font-weight-normal">{status.status.title}</h2>
        <p>Checks code using ESLint, an open-source JavaScript linting utility.</p>
        <Timeline className="align-items-stretch mb-3">
            <div className="d-flex align-items-start bg-body border p-3 mb-5">
                <EyeIcon className="icon-inline mb-0 mr-3" />
                Checking all repositories
            </div>
            <div className="d-flex align-items-start bg-body border p-3 mb-5">
                <SettingsIcon className="icon-inline mr-3 flex-0" />
                <ul className="list-unstyled mb-0 flex-1">
                    <li>
                        Use <code>eslint@6.0.1</code>
                    </li>
                    <li>Check for new, recommended ESLint rules</li>
                    <li>Ignore projects with only JavaScript files</li>
                </ul>
                <button className="btn btn-sm btn-secondary mb-3 flex-0">Configure</button>
            </div>
            <div className="d-flex align-items-start bg-body border p-3 mb-5">
                <BellIcon className="icon-inline mr-3 flex-0" />
                <ul className="list-unstyled mb-0 flex-1">
                    <li>Fail changesets that add code not checked by ESLint</li>
                    <li>
                        Notify <strong>@felixfbecker</strong> of new ESLint rules
                    </li>
                </ul>
                <StatusNotificationSettingsDropdownButton buttonClassName="btn-secondary btn-sm flex-0" />
            </div>
            <StatusStateBar status={status} className="p-3 bg-body" />
        </Timeline>
    </div>
)
