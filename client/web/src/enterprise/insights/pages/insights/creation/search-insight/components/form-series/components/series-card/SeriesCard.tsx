import classNames from 'classnames'
import React, { ReactElement } from 'react'

import { Button } from '@sourcegraph/wildcard'

import { DEFAULT_ACTIVE_COLOR } from '../../../form-color-input/FormColorInput'

import styles from './SeriesCard.module.scss'

interface SeriesCardProps {
    isRemoveSeriesAvailable: boolean

    /** Name of series. */
    name: string
    /** Query value of series. */
    query: string
    /** Color value of series. */
    stroke?: string
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
    const {
        isRemoveSeriesAvailable,
        name,
        query,
        stroke: color = DEFAULT_ACTIVE_COLOR,
        className,
        onEdit,
        onRemove,
    } = props

    return (
        <li
            data-testid="series-card"
            aria-label={`${name} data series`}
            className={classNames(styles.card, className, 'card d-flex flex-row p-3')}
        >
            <div className={styles.cardInfo}>
                <div className={classNames('mb-1 ', styles.cardTitle)}>
                    {/* eslint-disable-next-line react/forbid-dom-props */}
                    <div data-testid="series-color-mark" style={{ color }} className={styles.cardColorMark} />
                    <span
                        data-testid="series-name"
                        title={name}
                        className={classNames(styles.cardName, 'ml-1 font-weight-bold')}
                    >
                        {name}
                    </span>
                </div>

                <span data-testid="series-query" className="mb-0 text-muted">
                    {query}
                </span>
            </div>

            <div className="d-flex align-items-center">
                <Button
                    data-testid="series-edit-button"
                    type="button"
                    onClick={onEdit}
                    variant="primary"
                    outline={true}
                    className="border-0"
                >
                    Edit
                </Button>

                <Button
                    data-testid="series-delete-button"
                    type="button"
                    onClick={onRemove}
                    disabled={!isRemoveSeriesAvailable}
                    className="border-0 ml-1"
                    variant="danger"
                    outline={true}
                >
                    Remove
                </Button>
            </div>
        </li>
    )
}
