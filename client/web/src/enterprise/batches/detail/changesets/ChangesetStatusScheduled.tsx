import React, { useCallback, useMemo, useState } from 'react'
import TimerOutlineIcon from 'mdi-react/TimerOutlineIcon'
import classNames from 'classnames'
import { getChangesetScheduleEstimate } from '../backend'
import { formatDistance, isBefore, parseISO } from 'date-fns'

// This is copied from ChangesetStatusCell.
const iconClassNames = 'm-0 text-nowrap flex-column align-items-center justify-content-center'

// The world's smallest state machine: Date means we have an estimate; 'initial'
// is the initial state (gasp); 'loading' means we're waiting for the backend;
// null means the backend couldn't provide an estimate (which, practically
// speaking, means there are either no rollout windows configured or the
// estimate is more than a week away).
type MemoisedEstimate = Date | 'initial' | 'loading' | null

const estimateTooltip = (estimate: MemoisedEstimate) => {
    if (estimate === 'initial' || estimate === 'loading') {
        return null
    }

    if (estimate) {
        const now = new Date()
        if (isBefore(estimate, now)) {
            return 'This changeset will be processed soon.'
        } else {
            return `This changeset will be processed in approximately ${formatDistance(estimate, now)}.`
        }
    }

    return 'No estimate is available for when this changeset will be processed.'
}

export const ChangesetStatusScheduled: React.FunctionComponent<{
    id: string
    label?: JSX.Element
    className?: string
}> = ({ id, label = <span>Scheduled</span>, className }) => {
    // Calculating the estimate is just expensive enough that we don't want to
    // do it for every changeset. (If we did, we'd just request the field when
    // we make the initial GraphQL call to list the changesets.)
    //
    // As a result, we only trigger the initial load of the estimated processing
    // time when the user mouses over the status component for the first time.
    // After that, we'll cache it: this isn't a value that's likely to change
    // very much, and when the changeset is processed, this component is going
    // to be replaced by a different one anyway.

    const [estimate, setEstimate] = useState<MemoisedEstimate>('initial')
    const [tooltip, setTooltip] = useState<string | null>(null)
    const onMouseOver = useCallback(async () => {
        if (estimate === 'initial') {
            // Initially, there was a loading state in the tooltip, but updating
            // the tooltip text with a stationary cursor is honestly pretty
            // janky, so it's better to minimise the number of updates.
            //
            // (We could use Tooltip.forceUpdate() in theory, but it doesn't
            // play very nicely with keeping the tooltip in a state variable in
            // practice. It doesn't feel worth the hassle.)
            setEstimate('loading')
            const raw = await getChangesetScheduleEstimate(id)
            if (raw) {
                setEstimate(parseISO(raw))
                setTooltip(estimateTooltip(estimate))
            } else {
                setEstimate(null)
            }
        } else if (estimate !== 'loading' && estimate !== null) {
            // If we already have an estimate, then we should update the
            // tooltip, since it has a relative time.
            setTooltip(estimateTooltip(estimate))
        }
    }, [estimate, id])

    return (
        <div className={classNames(iconClassNames, className)} onMouseOver={onMouseOver} data-tooltip={tooltip}>
            <TimerOutlineIcon />
            {label}
        </div>
    )
}
