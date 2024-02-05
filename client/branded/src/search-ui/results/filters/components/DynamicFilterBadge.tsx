import { FC } from 'react'

import classNames from 'classnames'

import { Badge, Tooltip } from '@sourcegraph/wildcard'

import styles from './DynamicFilterBadge.module.scss'

export const DynamicFilterBadge: FC<{ exhaustive: boolean; count: number }> = ({ exhaustive, count }) => {
    const tooltipContent = exhaustive ? null : (
        <>
            This is an approximate count of the results returned because you hit a limit. Try increasing the limit using
            the <code>count:</code> filter in the search query, or select <code>count:all</code> from the filter list.
        </>
    )

    return (
        <Tooltip content={tooltipContent} placement="right">
            <Badge ref={null} variant="secondary" className={classNames('ml-2', styles.countBadge)}>
                {exhaustive ? count : `${roundCount(count)}+`}
            </Badge>
        </Tooltip>
    )
}

function roundCount(count: number): number {
    const roundNumbers = [10000, 5000, 1000, 500, 100, 50, 10, 5, 1]
    for (const roundNumber of roundNumbers) {
        if (count >= roundNumber) {
            return roundNumber
        }
    }
    return 0
}
