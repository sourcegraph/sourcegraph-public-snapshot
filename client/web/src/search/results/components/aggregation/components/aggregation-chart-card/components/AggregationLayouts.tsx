import { type FC, forwardRef, type PropsWithChildren, type SVGProps } from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import styles from './AggregationLayouts.module.scss'

interface AggregationTextContentProps {
    size: 'sm' | 'md'
    className?: string
}

export const AggregationTextContent: FC<PropsWithChildren<AggregationTextContentProps>> = props => {
    const { size, className, children } = props

    return (
        <AggregationContent size={size} data-error-layout={true} className={classNames(styles.textLayout, className)}>
            <BarsBackground size={size} />

            {children && (
                <div className={styles.textLayoutContainer}>
                    <div className={styles.textLayoutMessage}>{children}</div>
                </div>
            )}
        </AggregationContent>
    )
}

export const AggregationContent = forwardRef(function DataLayoutContainerRef(props, ref) {
    const { as: Component = 'div', size = 'md', className, ...attributes } = props

    return (
        <Component
            {...attributes}
            ref={ref}
            className={classNames(className, styles.layout, {
                [styles.layoutSmall]: size === 'sm',
            })}
        />
    )
}) as ForwardReferenceComponent<'div', { size?: 'sm' | 'md' }>

const BAR_VALUES_FULL_UI = [95, 88, 83, 70, 65, 45, 35, 30, 30, 30, 30, 27, 27, 27, 27, 24, 10, 10, 10, 10, 10]
const BAR_VALUES_SIDEBAR_UI = [95, 80, 75, 68, 62, 55, 45, 45, 40, 40, 38, 33, 30, 25, 15, 7]

interface BarsBackgroundProps extends SVGProps<SVGSVGElement> {
    size: 'sm' | 'md'
}

const BarsBackground: FC<BarsBackgroundProps> = props => {
    const { size, className, ...attributes } = props

    const padding = size === 'md' ? 1 : 2
    const data = size === 'md' ? BAR_VALUES_FULL_UI : BAR_VALUES_SIDEBAR_UI
    const barWidth = (100 - padding * (data.length - 1)) / data.length

    return (
        <svg
            role="presentation"
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
