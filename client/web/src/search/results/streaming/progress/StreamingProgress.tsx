import * as React from 'react'
import { Progress } from '../../../stream'
import { StreamingProgressCount } from './StreamingProgressCount'
import { StreamingProgressSkippedButton } from './StreamingProgressSkippedButton'

export interface StreamingProgressProps {
    progress?: Progress
    onSearchAgain?: (additionalFilters: string[]) => void
}

export const defaultProgress: Progress = {
    done: true,
    durationMs: 0,
    matchCount: 0,
    skipped: [],
}

export const StreamingProgress: React.FunctionComponent<StreamingProgressProps> = props => (
    <div className="d-flex streaming-progress">
        <StreamingProgressCount {...props} />
        <StreamingProgressSkippedButton {...props} />
    </div>
)
