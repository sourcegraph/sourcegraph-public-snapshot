import classnames from 'classnames'
import React, { ReactElement } from 'react'

import styles from './SeriesCard.module.scss'

interface SeriesCardProps {
    /** Name of series. */
    name: string
    /** Query value of series. */
    query: string
    /** Color value of series. */
    stroke: string
    /** Custom class name for root button element. */
    className?: string
    /** Edit handler. */
    onEdit?: () => void
    /** Remove handler. */
    onRemove?: () => void
}

/**
 * Renders series card component, visual list item of series (name, color, query)
 * */
export function SeriesCard(props: SeriesCardProps): ReactElement {
    const { name, query, stroke: color, className, onEdit, onRemove } = props

    return (
        <li
            aria-label={`${name} data series`}
            className={classnames(styles.card, className, 'card d-flex flex-row p-3')}
        >
            <div className={styles.cardInfo}>
                <div className={classnames('mb-1 ', styles.cardTitle)}>
                    {/* eslint-disable-next-line react/forbid-dom-props */}
                    <div style={{ color }} className={styles.cardColorMark} />
                    <span title={name} className={classnames(styles.cardName, 'ml-1 font-weight-bold')}>
                        {name}
                    </span>
                </div>

                <span className="mb-0 text-muted">{query}</span>
            </div>

            <div className="d-flex align-items-center">
                <button type="button" onClick={onEdit} className="border-0 btn btn-outline-primary">
                    Edit
                </button>

                <button type="button" onClick={onRemove} className="border-0 btn btn-outline-danger ml-1">
                    Remove
                </button>
            </div>
        </li>
    )
}
