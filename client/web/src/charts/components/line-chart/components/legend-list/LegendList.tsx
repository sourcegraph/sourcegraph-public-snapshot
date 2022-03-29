import React from 'react'

import classNames from 'classnames'

import styles from './LegendList.module.scss'

interface LegendListProps extends React.HTMLAttributes<HTMLUListElement> {
    className?: string
}

export const LegendList: React.FunctionComponent<LegendListProps> = props => {
    const { className, ...attributes } = props

    return (
        <ul {...attributes} className={classNames(styles.legendList, className)}>
            {props.children}
        </ul>
    )
}

interface LegendItemProps {
    color: string
    name: string
}

export const LegendItem: React.FunctionComponent<LegendItemProps> = props => (
    <li className={styles.legendItem}>
        <div
            /* eslint-disable-next-line react/forbid-dom-props */
            style={{ backgroundColor: props.color }}
            className={styles.legendMark}
        />
        {props.name}
    </li>
)
