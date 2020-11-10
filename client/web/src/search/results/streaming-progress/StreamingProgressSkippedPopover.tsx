import * as React from 'react'
import { defaultProgress, StreamingProgressProps } from './StreamingProgress'

export const StreamingProgressSkippedPopover: React.FunctionComponent<StreamingProgressProps> = ({
    progress = defaultProgress,
}) => <div>Hello world</div>
