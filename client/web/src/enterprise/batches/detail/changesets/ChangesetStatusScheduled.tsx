import React, { useCallback, useState } from 'react'

import { mdiTimerOutline } from '@mdi/js'
import classNames from 'classnames'
import { formatDistanceToNow, isBefore, parseISO } from 'date-fns'

import type { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { Tooltip, Icon } from '@sourcegraph/wildcard'

import { getChangesetScheduleEstimate } from '../backend'

// This is copied from ChangesetStatusCell.
const iconClassNames = 'm-0 text-nowrap flex-column align-items-center justify-content-center'

// The world's smallest state machine: Date means we have an estimate; 'initial'
// is the initial state (gasp); 'loading' means we're waiting for the backend;
// null means the backend couldn't provide an estimate (which, practically
// speaking, means there are either no rollout windows configured or the
// estimate is more than a week away).
type MemoisedEstimate = Date | 'initial' | 'loading' | null

const estimateTooltip = (estimate: MemoisedEstimate): string | null => {
    if (estimate === 'initial' || estimate === 'loading') {
        return null
    }

    if (estimate) {
        const now = new Date()
        if (isBefore(estimate, now)) {
            return 'This changeset will be processed soon.'
        }
        // formatDistanceToNow() usually includes modifiers like "about" for
        // hazier date ranges, so we don't need to hedge here in the static
        // verbiage.
        return `This changeset will be processed in ${formatDistanceToNow(estimate)}.`
    }

    return 'No estimate is available for when this changeset will be processed.'
}

interface Props {
    id: Scalars['ID']
    label: JSX.Element
    className?: string
    role?: React.AriaRole
}

const DynamicChangesetStatusScheduled: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    id,
    label,
    className,
    role,
}) => {
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
        <Tooltip content={tooltip}>
            <div
                className={classNames(iconClassNames, className)}
                onMouseOver={onMouseOver}
                onFocus={onMouseOver}
                role={role}
            >
                <Icon svgPath={mdiTimerOutline} inline={false} aria-hidden={true} />
                {label}
            </div>
        </Tooltip>
    )
}

const StaticChangesetStatusScheduled: React.FunctionComponent<
    React.PropsWithChildren<Pick<Props, 'label' | 'className' | 'role'>>
> = ({ label, className, role }) => (
    <div className={classNames(iconClassNames, className)} role={role}>
        <Icon svgPath={mdiTimerOutline} inline={false} aria-hidden={true} />
        {label}
    </div>
)

export const ChangesetStatusScheduled: React.FunctionComponent<React.PropsWithChildren<Partial<Props>>> = ({
    id,
    label = <span>Scheduled</span>,
    className,
    role,
}) => (
    // If there's no ID (for example, when previewing a batch change), then no
    // dynamic behaviour is required, and we can just return a static icon and
    // label. Otherwise, we need the whole dynamic shebang.
    <>
        {id ? (
            <DynamicChangesetStatusScheduled id={id} label={label} className={className} role={role} />
        ) : (
            <StaticChangesetStatusScheduled label={label} className={className} role={role} />
        )}
    </>
)
