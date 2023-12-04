import type { FunctionComponent } from 'react'

import { intervalToDuration } from 'date-fns'
import formatDuration from 'date-fns/formatDuration'

import { Tooltip } from '@sourcegraph/wildcard'

export interface DurationProps {
    hours: number
}

const MS_IN_HOURS = 1000 * 60 * 60

export const Duration: FunctionComponent<DurationProps> = ({ hours }) => (
    <Tooltip content={`${hours} hours`}>
        <span className="timestamp">{formatDuration(intervalToDuration({ start: 0, end: hours * MS_IN_HOURS }))}</span>
    </Tooltip>
)
