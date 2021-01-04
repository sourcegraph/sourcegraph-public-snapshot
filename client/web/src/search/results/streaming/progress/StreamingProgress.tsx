import * as React from 'react'
import { Progress, StreamingResultsState } from '../../../stream'
import { StreamingProgressCount } from './StreamingProgressCount'
import { StreamingProgressSkippedButton } from './StreamingProgressSkippedButton'

export interface StreamingProgressProps {
    state: StreamingResultsState
    progress: Progress
    onSearchAgain: (additionalFilters: string[]) => void
}

export const StreamingProgress: React.FunctionComponent<StreamingProgressProps> = props => (
    <div className="d-flex align-items-center streaming-progress">
        <StreamingProgressCount {...props} />
        <StreamingProgressSkippedButton {...props} />
    </div>
)
