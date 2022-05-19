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
}

export const LegendItem: React.FunctionComponent<React.PropsWithChildren<LegendItemProps>> = ({
    color,
    name,
    className,
    ...attributes
}) => (
    <li {...attributes} className={classNames(styles.legendItem, className)}>
        <span
            /* eslint-disable-next-line react/forbid-dom-props */
            style={{ backgroundColor: color }}
            className={styles.legendMark}
        />
        {name}
    </li>
)
