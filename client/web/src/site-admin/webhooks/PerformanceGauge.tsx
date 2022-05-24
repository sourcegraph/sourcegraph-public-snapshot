import React from 'react'

import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'

import styles from './PerformanceGauge.module.scss'

export interface Props {
    count?: number
    label: string
    plural?: string

    className?: string
    countClassName?: string
    labelClassName?: string
}

/**
 * A performance gauge is a component that renders a numeric value with a label
 * in a way that focuses attention on the numeric value.
 */
export const PerformanceGauge: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    count,
    className,
    countClassName,
    label,
    labelClassName,
    plural,
}) => (
    <div className={classNames(styles.gauge, 'px-4', 'py-3', 'd-flex', 'align-items-center', className)}>
        {count === undefined ? (
            <span className={classNames(styles.count, 'text-muted', countClassName)}>&hellip;</span>
        ) : (
            <span className={classNames(styles.count, countClassName)}>{count}</span>
        )}
        <span className={classNames('text-muted', labelClassName)}>{pluralize(label, count ?? 0, plural)}</span>
    </div>
)
