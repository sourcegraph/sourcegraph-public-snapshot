import { FC } from 'react'

import { mdiAlertCircle } from '@mdi/js'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

interface AggregationLimitLabelProps {
    size: 'sm' | 'md'
}

export const AggregationLimitLabel: FC<AggregationLimitLabelProps> = props => {
    const { size } = props

    const Component = size === 'sm' ? 'small' : 'span'

    return (
        <Tooltip content="This search exceeded the grouping limit. Results may be incomplete.">
            <Component className="text-muted">
                <Icon color="var(--warning)" svgPath={mdiAlertCircle} aria-hidden={true} className="mr-1" />
                Group limit reached
            </Component>
        </Tooltip>
    )
}
