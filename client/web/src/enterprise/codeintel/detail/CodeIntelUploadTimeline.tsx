import CheckIcon from 'mdi-react/CheckIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import FileUploadIcon from 'mdi-react/FileUploadIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import React, { FunctionComponent, useMemo } from 'react'
import { Maybe } from '../../../../../shared/src/graphql-operations'
import { Timeline, TimelineStage } from '../../../components/Timeline'
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
}) => {
    const stages = useMemo(() => {
        const stages = [...dumpUploadStage(upload), ...processingStage(upload)]
        if (upload.state === LSIFUploadState.COMPLETED) {
            stages.push({ icon: <CheckIcon />, text: 'Finished', date: upload.finishedAt, className: 'bg-success' })
        }
        return stages
    }, [upload])
    return <Timeline stages={stages} now={now} className={className} />
}

const dumpUploadStage = (upload: LsifUploadFields): TimelineStage[] => {
    const stage: TimelineStage[] = [
        {
            icon: <FileUploadIcon />,
            text: 'Upload started',
            date: upload.uploadedAt,
            className: upload.startedAt !== null ? 'bg-success' : 'bg-danger',
        },
    ]

    if (upload.startedAt === null) {
        // not 100% accurate, as we don't store the time of failure, only the start time of upload
        stage.push(failedStage(upload.uploadedAt))
    }

    return stage
}

const processingStage = (upload: LsifUploadFields): TimelineStage[] => {
    const stage: TimelineStage[] = [
        {
            icon: <ProgressClockIcon />,
            text: 'Began processing',
            date: upload.startedAt,
            className: 'bg-success',
        },
    ]

    if (upload.state === LSIFUploadState.ERRORED) {
        // ok because finishedAt is set when processing fails
        stage.push(failedStage(upload.finishedAt))
    }

    return stage
}

const failedStage = (date: Maybe<string>): TimelineStage => ({
    icon: <ErrorIcon />,
    text: 'Failed',
    date,
    className: 'bg-danger',
})
