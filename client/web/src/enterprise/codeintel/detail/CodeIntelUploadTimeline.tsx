import CheckIcon from 'mdi-react/CheckIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import FileUploadIcon from 'mdi-react/FileUploadIcon'
import ProgressClockIcon from 'mdi-react/ProgressClockIcon'
import React, { FunctionComponent, useMemo } from 'react'

import { LSIFUploadState } from '@sourcegraph/shared/src/graphql-operations'
import { Timeline, TimelineStage } from '@sourcegraph/web/src/components/Timeline'

import { LsifUploadFields } from '../../../graphql-operations'

export interface CodeIntelUploadTimelineProps {
    upload: LsifUploadFields
    now?: () => Date
    className?: string
}

enum FailedStage {
    UPLOADING,
    PROCESSING,
}

export const CodeIntelUploadTimeline: FunctionComponent<CodeIntelUploadTimelineProps> = ({
    upload,
    now,
    className,
}) => {
    let failedStage: FailedStage | null = null
    if (upload.state === LSIFUploadState.ERRORED && upload.startedAt === null) {
        failedStage = FailedStage.UPLOADING
    } else if (upload.state === LSIFUploadState.ERRORED && upload.startedAt !== null) {
        failedStage = FailedStage.PROCESSING
    }

    const stages = useMemo(
        () =>
            [uploadStages, processingStages, terminalStages].flatMap(stageConstructor =>
                stageConstructor(upload, failedStage)
            ),
        [upload, failedStage]
    )

    return <Timeline stages={stages} now={now} className={className} />
}

const uploadStages = (upload: LsifUploadFields, failedStage: FailedStage | null): TimelineStage[] => [
    {
        icon: <FileUploadIcon />,
        text:
            upload.state === LSIFUploadState.UPLOADING ||
            (LSIFUploadState.ERRORED && failedStage === FailedStage.UPLOADING)
                ? 'Upload started'
                : 'Uploaded',
        date: upload.uploadedAt,
        className:
            upload.state === LSIFUploadState.UPLOADING
                ? 'bg-primary'
                : upload.state === LSIFUploadState.ERRORED
                ? failedStage === FailedStage.UPLOADING
                    ? 'bg-danger'
                    : 'bg-success'
                : 'bg-success',
    },
]

const processingStages = (upload: LsifUploadFields, failedStage: FailedStage | null): TimelineStage[] => [
    {
        icon: <ProgressClockIcon />,
        text:
            upload.state === LSIFUploadState.PROCESSING ||
            (LSIFUploadState.ERRORED && failedStage === FailedStage.PROCESSING)
                ? 'Processing started'
                : 'Processed',
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
