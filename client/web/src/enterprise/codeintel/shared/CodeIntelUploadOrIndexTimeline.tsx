import CheckIcon from 'mdi-react/CheckIcon'
import ClockFastIcon from 'mdi-react/ClockFastIcon'
import ClockStartIcon from 'mdi-react/ClockStartIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { FunctionComponent } from 'react'
import { Timeline } from '../../../components/Timeline'
import { LsifIndexFields, LSIFIndexState, LsifUploadFields, LSIFUploadState } from '../../../graphql-operations'

export interface TimelineNode {
    state?: LsifIndexFields['state'] | LsifUploadFields['state']
    queuedAt?: string | null
    uploadedAt?: string | null
    startedAt?: string | null
    finishedAt?: string | null
}

const isCompleted = (node: TimelineNode): boolean =>
    node.state === LSIFUploadState.COMPLETED || node.state === LSIFIndexState.COMPLETED

export interface CodeIntelUploadOrIndexTimelineProps {
    node: TimelineNode
    now?: () => Date
    className?: string
}

export const CodeIntelUploadOrIndexTimeline: FunctionComponent<CodeIntelUploadOrIndexTimelineProps> = ({
    node,
    now,
    className,
}) => (
    <Timeline
        stages={[
            { icon: <ClockStartIcon />, text: 'Queued', date: node.queuedAt, className: 'success' },
            { icon: <ClockStartIcon />, text: 'Uploaded', date: node.uploadedAt, className: 'success' },
            { icon: <ClockFastIcon />, text: 'Began processing', date: node.startedAt, className: 'success' },
            {
                icon: isCompleted(node) ? <CheckIcon /> : <CloseIcon />,
                text: isCompleted(node) ? 'Finished' : 'Failed',
                date: node.finishedAt,
                className: isCompleted(node) ? 'success' : 'failure',
            },
        ]}
        now={now}
        className={className}
    />
)
