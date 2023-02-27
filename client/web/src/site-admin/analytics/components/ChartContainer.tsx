import React from 'react'

import { Text, ParentSize } from '@sourcegraph/wildcard'

import styles from './index.module.scss'

interface ChartContainerProps {
    className?: string
    title?: string
    labelX?: string
    labelY?: string
    children: (width: number) => React.ReactNode
}

export const ChartContainer: React.FunctionComponent<ChartContainerProps> = ({
    className,
    title,
    children,
    labelX,
    labelY,
}) => (
    <div className={className}>
        {title && <Text alignment="center">{title}</Text>}
        <div className="d-flex">
            {labelY && <span className={styles.chartYLabel}>{labelY}</span>}
            <ParentSize>{({ width }) => children(width)}</ParentSize>
        </div>
        {labelX && <div className={styles.chartXLabel}>{labelX}</div>}
    </div>
)
