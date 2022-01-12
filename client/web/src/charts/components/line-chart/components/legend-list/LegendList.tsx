import classNames from 'classnames'
import React from 'react'

import { LineChartSeries } from '../../types'
import { getLineColor } from '../../utils/colors'

import styles from './LegendList.module.scss'

interface LegendListProps {
    series: LineChartSeries<any>[]
    className?: string
}

export const LegendList: React.FunctionComponent<LegendListProps> = props => {
    const { series, className } = props

    return (
        <ul className={classNames(styles.legendList, className)}>
            {series.map(line => (
                <li key={line.dataKey.toString()} className={styles.legendItem}>
                    <div
                        /* eslint-disable-next-line react/forbid-dom-props */
                        style={{ backgroundColor: getLineColor(line) }}
                        className={styles.legendMark}
                    />
                    {line.name}
                </li>
            ))}
        </ul>
    )
}
