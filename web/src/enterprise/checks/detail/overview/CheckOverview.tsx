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

interface Props extends Pick<CheckAreaContext, 'checkID' | 'checkInfo' | 'checkURL' | 'checksURL'> {
    className?: string
}

/**
 * An overview of a check.
 */
export const CheckOverview: React.FunctionComponent<Props> = ({
    checkID,
    checkInfo,
    checkURL,
    checksURL,
    className = '',
}) => (
    <div className={`check-overview ${className || ''}`}>
        <CheckBreadcrumbs
            checkID={checkID}
            checkInfo={checkInfo}
            checkURL={checkURL}
            checksURL={checksURL}
            className="py-3"
        />
        <hr className="my-0" />
        <h2 className="my-3 font-weight-normal">
            {checkID.type} {checkID.id}
        </h2>
        {checkInfo.description && <Markdown dangerousInnerHTML={renderMarkdown(checkInfo.description.value)} />}
        <Timeline tag="div" className="align-items-stretch mb-3">
            <div className="d-flex align-items-start bg-body border p-3 mb-5">
                <EyeIcon className="icon-inline mb-0 mr-3" />
                Checking all repositories
            </div>
            {checkInfo.sections && (
                <>
                    {checkInfo.sections.settings && (
                        <CheckPipelineSection
                            section="settings"
                            content={checkInfo.sections.settings}
                            // tslint:disable-next-line: jsx-no-lambda
                            action={className => (
                                <Link to={`${checkURL}/settings`} className={`btn ${className}`}>
                                    Configure
                                </Link>
                            )}
                        />
                    )}
                    {checkInfo.sections.notifications && (
                        <CheckPipelineSection
                            section="notifications"
                            content={checkInfo.sections.notifications}
                            // tslint:disable-next-line: jsx-no-lambda
                            action={className => (
                                <CheckNotificationSettingsDropdownButton buttonClassName={className} />
                            )}
                        />
                    )}
                </>
            )}
            <CheckStateBar checkInfo={checkInfo} className="p-3 bg-body" />
        </Timeline>
    </div>
)
