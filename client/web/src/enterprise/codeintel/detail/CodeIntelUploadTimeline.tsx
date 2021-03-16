import CheckIcon from 'mdi-react/CheckIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import FileUploadIcon from 'mdi-react/FileUploadIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import React, { FunctionComponent, useMemo } from 'react'
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
    const stages = useMemo(
        () => [uploadStages, processingStages, terminalStages].flatMap(stageConstructor => stageConstructor(upload)),
        [upload]
    )

    return <Timeline stages={stages} now={now} className={className} />
}

const uploadStages = (upload: LsifUploadFields): TimelineStage[] => [
    {
        icon: <FileUploadIcon />,
        text: upload.state === LSIFUploadState.UPLOADING ? 'Upload started' : 'Uploaded',
        date: upload.uploadedAt,
        className:
            upload.state === LSIFUploadState.UPLOADING
                ? 'bg-primary'
                : upload.state === LSIFUploadState.ERRORED
                ? 'bg-danger'
                : 'bg-success',
    },
]

const processingStages = (upload: LsifUploadFields): TimelineStage[] => [
    {
        icon: <ProgressClockIcon />,
        text: upload.state === LSIFUploadState.PROCESSING ? 'Processing started' : 'Processed',
        date: upload.startedAt,
        className:
            upload.state === LSIFUploadState.PROCESSING
                ? 'bg-primary'
                : upload.state === LSIFUploadState.ERRORED
                ? 'bg-danger'
                : 'bg-success',
    },
]

const terminalStages = (upload: LsifUploadFields): TimelineStage[] =>
    upload.state === LSIFUploadState.COMPLETED
        ? [
              {
                  icon: <CheckIcon />,
                  text: 'Finished',
                  date: upload.finishedAt,
                  className: 'bg-success',
              },
          ]
        : upload.state === LSIFUploadState.ERRORED
        ? [
              {
                  icon: <ErrorIcon />,
                  text: 'Failed',
                  date: upload.finishedAt,
                  className: 'bg-danger',
              },
          ]
        : []
