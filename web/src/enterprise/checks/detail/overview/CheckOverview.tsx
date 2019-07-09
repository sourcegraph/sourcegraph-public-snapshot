import EyeIcon from 'mdi-react/EyeIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { Markdown } from '../../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../../shared/src/util/markdown'
import { Timeline } from '../../../../components/timeline/Timeline'
import { CheckAreaContext } from '../CheckArea'
import { CheckBreadcrumbs } from './CheckBreadcrumbs'
import { CheckPipelineSection } from './pipeline/CheckPipelineSection'
import { CheckNotificationSettingsDropdownButton } from './stateBar/CheckNotificationSettingsDropdownButton'
import { CheckStateBar } from './stateBar/CheckStateBar'

interface Props extends Pick<CheckAreaContext, 'status' | 'statusURL' | 'statusesURL'> {
    className?: string
}

/**
 * An overview of a status.
 */
export const CheckOverview: React.FunctionComponent<Props> = ({ status, statusURL, statusesURL, className = '' }) => (
    <div className={`status-overview ${className || ''}`}>
        <CheckBreadcrumbs status={status} statusURL={statusURL} statusesURL={statusesURL} className="py-3" />
        <hr className="my-0" />
        <h2 className="my-3 font-weight-normal">{status.status.title}</h2>
        {status.status.description && <Markdown dangerousInnerHTML={renderMarkdown(status.status.description.value)} />}
        <Timeline tag="div" className="align-items-stretch mb-3">
            <div className="d-flex align-items-start bg-body border p-3 mb-5">
                <EyeIcon className="icon-inline mb-0 mr-3" />
                Checking all repositories
            </div>
            {status.status.sections && (
                <>
                    {status.status.sections.settings && (
                        <CheckPipelineSection
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
                        <CheckPipelineSection
                            section="notifications"
                            content={status.status.sections.notifications}
                            // tslint:disable-next-line: jsx-no-lambda
                            action={className => (
                                <CheckNotificationSettingsDropdownButton buttonClassName={className} />
                            )}
                        />
                    )}
                </>
            )}
            <CheckStateBar status={status} className="p-3 bg-body" />
        </Timeline>
    </div>
)
