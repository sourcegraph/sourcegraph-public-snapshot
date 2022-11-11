import { forwardRef } from 'react'

import classNames from 'classnames'

import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import styles from './LegendList.module.scss'

export const LegendList = forwardRef(function LegendList(props, ref) {
    const { as: Component = 'ul', 'aria-label': ariaLabel = 'Chart legend', className, ...attributes } = props

    return (
        <Component
            {...attributes}
            ref={ref}
            aria-label={ariaLabel}
            className={classNames(styles.legendList, className)}
        />
    )
}) as ForwardReferenceComponent<'ul'>

interface LegendItemProps {
    name: string
    hovered?: boolean
    selected?: boolean
    color?: string
}

export const LegendItem = forwardRef(function LegendItem(props, ref) {
    const {
        as: Component = 'span',
        name,
        hovered,
        selected = true,
        color = 'var(--gray-07)',
        className,
        children,
        ...attributes
    } = props

    return (
        <li ref={ref}>
            <Component
                {...attributes}
                className={classNames(styles.legendItem, className, { 'text-muted': !selected && !hovered })}
            >
                <span
                    aria-hidden={true}
                    /* eslint-disable-next-line react/forbid-dom-props */
                    style={{ backgroundColor: selected || hovered ? color : undefined }}
                    className={classNames([styles.legendMark, { [styles.unselected]: !selected }])}
                />
                {children || name}
            </Component>
        </li>

    )
}) as ForwardReferenceComponent<'li', LegendItemProps>
