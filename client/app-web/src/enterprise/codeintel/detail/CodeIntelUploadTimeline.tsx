import CheckIcon from 'mdi-react/CheckIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import FileUploadIcon from 'mdi-react/FileUploadIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import React, { FunctionComponent } from 'react'
import { Timeline } from '../../../components/Timeline'
import { LsifUploadFields, LSIFUploadState } from '../../../graphql-operations'

export interface CodeIntelUploadTimelineProps {
    upload: LsifUploadFields
    now?: () => Date
    className?: string
}

export const CodeIntelUploadTimeline: FunctionComponent<CodeIntelUploadTimelineProps> = ({
    upload,
    now,
    className,
}) => (
    <Timeline
        stages={[
            { icon: <FileUploadIcon />, text: 'Uploaded', date: upload.uploadedAt, className: 'bg-success' },
            { icon: <ProgressClockIcon />, text: 'Began processing', date: upload.startedAt, className: 'bg-success' },
            upload.state === LSIFUploadState.COMPLETED
                ? { icon: <CheckIcon />, text: 'Finished', date: upload.finishedAt, className: 'bg-success' }
                : { icon: <ErrorIcon />, text: 'Failed', date: upload.finishedAt, className: 'bg-danger' },
        ]}
        now={now}
        className={className}
    />
)
