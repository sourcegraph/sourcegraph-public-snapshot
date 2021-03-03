import React from 'react'
import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import { CampaignFields } from '../../../graphql-operations'

interface ClosedNoticeProps {
    closedAt: CampaignFields['closedAt']
    className?: string
}

export const ClosedNotice: React.FunctionComponent<ClosedNoticeProps> = ({ closedAt, className }) => {
    if (!closedAt) {
        return <></>
    }
    return (
        <div className={classNames('alert alert-info', className)}>
            <InformationOutlineIcon className="icon-inline" /> Information on this page may be out of date because
            changesets that only exist in closed campaigns are not synced with the code host.
        </div>
    )
}
