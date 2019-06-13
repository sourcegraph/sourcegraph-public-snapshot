import React from 'react'
import { Link } from 'react-router-dom'
import { Markdown } from '../../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../../shared/src/util/markdown'
import { Timeline } from '../../../../components/timeline/Timeline'
import { CheckAreaContext } from '../CheckArea'
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
        {checkInfo.description && <Markdown dangerousInnerHTML={renderMarkdown(checkInfo.description.value)} />}
        <Timeline tag="div" className="align-items-stretch mb-3">
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
