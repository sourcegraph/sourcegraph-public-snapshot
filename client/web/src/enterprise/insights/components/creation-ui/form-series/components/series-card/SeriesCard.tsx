import type { ReactElement } from 'react'

import classNames from 'classnames'

import { Button, Card } from '@sourcegraph/wildcard'

import { DEFAULT_DATA_SERIES_COLOR } from '../../../../../constants'

import styles from './SeriesCard.module.scss'

interface SeriesCardProps {
    disabled: boolean

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
    const { disabled, name, query, stroke: color = DEFAULT_DATA_SERIES_COLOR, className, onEdit, onRemove } = props

    return (
        <Card
            as="li"
            data-testid="series-card"
            aria-label={`${name} data series`}
            aria-disabled={disabled}
            className={classNames(styles.card, className, { [styles.cardDisabled]: disabled })}
        >
            <div className={styles.cardInfo}>
                <div className={classNames('mb-1 ', styles.cardTitle)}>
                    <div
                        data-testid="series-color-mark"
                        /* eslint-disable-next-line react/forbid-dom-props */
                        style={{ color: disabled ? 'var(--icon-muted)' : color }}
                        className={styles.cardColorMark}
                    />
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
                    disabled={disabled}
                    className="border-0"
                >
                    Edit
                </Button>

                <Button
                    data-testid="series-delete-button"
                    type="button"
                    onClick={onRemove}
                    className="border-0 ml-1"
                    variant="danger"
                    outline={true}
                >
                    Remove
                </Button>
            </div>
        </Card>
    )
}
