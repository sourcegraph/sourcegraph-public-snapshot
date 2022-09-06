import { SVGProps, FC } from 'react'

import classNames from 'classnames'

import styles from './AggregationBarsBackground.module.scss'

const BAR_VALUES_FULL_UI = [95, 88, 83, 70, 65, 45, 35, 30, 30, 30, 30, 27, 27, 27, 27, 24, 10, 10, 10, 10, 10]
const BAR_VALUES_SIDEBAR_UI = [95, 80, 75, 70, 68, 68, 55, 40, 38, 33, 30, 25, 15, 7]

interface BarsBackgroundProps extends SVGProps<SVGSVGElement> {
    size: 'sm' | 'md'
}

export const BarsBackground: FC<BarsBackgroundProps> = props => {
    const { size, className, ...attributes } = props

    const padding = size === 'md' ? 1 : 2
    const data = size === 'md' ? BAR_VALUES_FULL_UI : BAR_VALUES_SIDEBAR_UI
    const barWidth = (100 - padding * (data.length - 1)) / data.length

    return (
        <svg
            {...attributes}
            className={classNames(className, styles.zeroStateBackground)}
            xmlns="http://www.w3.org/2000/svg"
        >
            {data.map((bar, index) => (
                <rect
                    key={index}
                    x={`${index * (barWidth + padding)}%`}
                    y={`${100 - bar}%`}
                    height={`${bar}%`}
                    width={`${barWidth}%`}
                />
            ))}
        </svg>
    )
}
