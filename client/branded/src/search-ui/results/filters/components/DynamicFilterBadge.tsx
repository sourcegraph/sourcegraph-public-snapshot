import { FC } from 'react'

import classNames from 'classnames'

import { Badge } from '@sourcegraph/wildcard'

import styles from './DynamicFilterBadge.module.scss'

export const DynamicFilterBadge: FC<{ exhaustive: boolean; count: number }> = ({ exhaustive, count }) => (
    <Badge variant="secondary" className={classNames('ml-2', styles.countBadge)}>
        {exhaustive ? count : `${roundCount(count)}+`}
    </Badge>
)

function roundCount(count: number): number {
    const roundNumbers = [10000, 5000, 1000, 500, 100, 50, 10, 5, 1]
    for (const roundNumber of roundNumbers) {
        if (count >= roundNumber) {
            return roundNumber
        }
    }
    return 0
}
