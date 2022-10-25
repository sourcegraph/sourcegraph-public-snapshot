import React, { LiHTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './LegendList.module.scss'

interface LegendListProps extends React.HTMLAttributes<HTMLUListElement> {
    className?: string
}

export const LegendList: React.FunctionComponent<React.PropsWithChildren<LegendListProps>> = props => {
    const { className, ...attributes } = props

    return (
        <ul {...attributes} className={classNames(styles.legendList, className)}>
            {props.children}
        </ul>
    )
}

interface LegendItemProps extends LiHTMLAttributes<HTMLLIElement> {
    name: string
    color?: string
    selected?: boolean
    hovered?: boolean
}

export const LegendItem: React.FunctionComponent<React.PropsWithChildren<LegendItemProps>> = ({
    color = 'var(--gray-07)',
    name,
    selected = true,
    hovered,
    className,
    children,
    ...attributes
}) => (
    <li {...attributes} className={classNames({ 'text-muted': !selected && !hovered }, styles.legendItem, className)}>
        <span
            aria-hidden={true}
            /* eslint-disable-next-line react/forbid-dom-props */
            style={{ backgroundColor: selected || hovered ? color : undefined }}
            className={classNames([styles.legendMark, { [styles.unselected]: !selected }])}
        />
        {children || name}
    </li>
)
