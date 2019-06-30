import EyeIcon from 'mdi-react/EyeIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { Timeline } from '../../../../components/timeline/Timeline'
import { StatusAreaContext } from '../StatusArea'
import { StatusPipelineSection } from './pipeline/StatusPipelineSection'
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
        <Timeline tag="div" className="align-items-stretch mb-3">
            <div className="d-flex align-items-start bg-body border p-3 mb-5">
                <EyeIcon className="icon-inline mb-0 mr-3" />
                Checking all repositories
            </div>
            {status.status.sections && (
                <>
                    {status.status.sections.settings && (
                        <StatusPipelineSection
                            section="settings"
                            content={status.status.sections.settings}
                            // tslint:disable-next-line: jsx-no-lambda
                            action={className => (
                                <Link to={`${statusURL}/settings`} className={`btn ${className}`}>
                                    Configure
                                </Link>
                            )}
                        />
                    )}
                    {status.status.sections.notifications && (
                        <StatusPipelineSection
                            section="notifications"
                            content={status.status.sections.notifications}
                            // tslint:disable-next-line: jsx-no-lambda
                            action={className => (
                                <StatusNotificationSettingsDropdownButton buttonClassName={className} />
                            )}
                        />
                    )}
                </>
            )}
            <StatusStateBar status={status} className="p-3 bg-body" />
        </Timeline>
    </div>
)
