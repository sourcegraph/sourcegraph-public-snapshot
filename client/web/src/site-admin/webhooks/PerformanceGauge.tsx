import classNames from 'classnames'
import React from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'

import styles from './PerformanceGauge.module.scss'

export interface Props {
    count?: number
    label: string
    plural?: string

    className?: string
    countClassName?: string
    labelClassName?: string
}

export const PerformanceGauge: React.FunctionComponent<Props> = ({
    count,
    className,
    countClassName,
    label,
    labelClassName,
    plural,
}) => (
    <div className={classNames(styles.gauge, 'p-3', 'd-flex', 'align-items-center', className)}>
        {count === undefined ? (
            <span className={classNames(styles.count, 'text-muted', countClassName)}>&hellip;</span>
        ) : (
            <span className={classNames(styles.count, countClassName)}>{count}</span>
        )}
        <span className={classNames('text-muted', labelClassName)}>{pluralize(label, count ?? 0, plural)}</span>
    </div>
)
