import * as React from 'react'
import { Progress } from '../../stream'
import { StreamingProgressCount } from './StreamingProgressCount'
import { StreamingProgressSkippedButton } from './StreamingProgressSkippedButton'

export interface StreamingProgressProps {
    progress: Progress
}

const defaultProgress: Progress = {
    done: true,
    durationMs: 0,
    matchCount: 0,
    skipped: [],
}

export const StreamingProgress: React.FunctionComponent<StreamingProgressProps> = ({ progress = defaultProgress }) => (
    <div className="d-flex streaming-progress">
        <StreamingProgressCount progress={progress} />
        <StreamingProgressSkippedButton progress={progress} />
    </div>
)
