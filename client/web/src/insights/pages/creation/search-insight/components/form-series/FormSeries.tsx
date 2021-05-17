import classnames from 'classnames'
import React, { ReactElement } from 'react'

import { DataSeries } from '../../../../../core/backend/types'
import { FormSeriesInput } from '../form-series-input/FormSeriesInput'

import styles from './FormSeries.module.scss'

export interface FormSeriesProps {
    /** Controlled value (series - chart lines) for series input component. */
    series?: DataSeries[]

    editSeries: (DataSeries | undefined)[]

    /** Change handler. */
    onChange: (series: DataSeries[]) => void

    /**
     * Live change series while user typing in active series form.
     * Used by consumers for getting latest values for live preview chart.
     * */
    onLiveChange: (liveSeries: DataSeries, isValid: boolean, index: number) => void

    onEditCard: (openedCardIndex: number) => void
    onEditCardClose: (closedCardIndex: number) => void
}

/**
 * Renders form series (sub-form) for series (chart lines) creation code insight form.
 * */
export const FormSeries: React.FunctionComponent<FormSeriesProps> = props => {
    const {
        series = [],
        editSeries,
        onChange,
        onEditCard,
        onEditCardClose,
        onLiveChange
    } = props

    // Add empty series form component that user be able to fill this form and create
    // another chart series.
    const handleAddSeries = (): void => {
        onEditCard(editSeries.length)
    }

    const handleEditSeriesCommit = (index: number, editedSeries: DataSeries): void => {
        const newSeries = [
            ...series.slice(0, index),
            editedSeries,
            ...series.slice(index + 1),
        ]

        onChange(newSeries)
        onEditCardClose(index)
    }

    const handleRequestToEdit = (index: number): void => {
        onEditCard(index)
    }

    const handleCancelEditSeries = (index: number): void => {
        onEditCardClose(index)
    }

    // In case if we don't have series we have to skip series list ui (components below)
    // and render simple series form component.
    if (series.length === 0) {
        return <FormSeriesInput
            autofocus={false}
            className="card card-body p-3"
            onSubmit={series => handleEditSeriesCommit(0, series)}
            onChange={(values, valid) => onLiveChange(values, valid, 0)}
        />
    }

    return (
        <div role="list" className="d-flex flex-column">
            {editSeries.map((line, index) =>
                editSeries[index] ? (
                    <FormSeriesInput
                        key={`${line?.name ?? ''}-${index}`}
                        cancel={true}
                        onSubmit={series => handleEditSeriesCommit(index, series)}
                        onCancel={() => handleCancelEditSeries(index)}
                        className={classnames('card card-body p-3', styles.formSeriesItem)}
                        onChange={(values, valid) => onLiveChange(values, valid, index)}
                        {...line}
                    />
                ) : (
                    series[index] &&
                        <SeriesCard
                            key={`${series[index].name}-${index}`}
                            onEdit={() => handleRequestToEdit(index)}
                            className={styles.formSeriesItem}
                            {...series[index]}
                        />
                )
            )}

            <button
                type="button"
                onClick={handleAddSeries}
                className={classnames(styles.formSeriesItem, styles.formSeriesAddButton, 'btn btn-link p-3')}
            >
                + Add another data series
            </button>
        </div>
    )
}

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
function SeriesCard(props: SeriesCardProps): ReactElement {
    const { name, query, stroke: color, className, onEdit } = props

    return (
        <button
            type="button"
            onClick={onEdit}
            aria-label={`Edit button for ${name} data series`}
            className={classnames(styles.formSeriesCard, className, 'd-flex p-3 btn btn-outline-secondary')}
        >
            <div className="flex-grow-1 d-flex flex-column align-items-start">
                <span className="mb-1 font-weight-bold">{name}</span>
                <span className="mb-0 text-muted">{query}</span>
            </div>

            {/* eslint-disable-next-line react/forbid-dom-props */}
            <div style={{ color }} className={styles.formSeriesCardColor} />
        </button>
    )
}
