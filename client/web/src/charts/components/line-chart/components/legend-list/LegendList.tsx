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
    color: string
    name: string
    selected?: boolean
}

export const LegendItem: React.FunctionComponent<React.PropsWithChildren<LegendItemProps>> = ({
    color,
    name,
    selected = true,
    className,
    ...attributes
}) => (
    <li {...attributes} className={classNames({ 'text-muted': !selected }, styles.legendItem, className)}>
        <span
            /* eslint-disable-next-line react/forbid-dom-props */
            style={{ backgroundColor: selected ? color : undefined }}
            className={classNames([styles.legendMark, { [styles.unselected]: !selected }])}
        />
        {name}
    </li>
)
