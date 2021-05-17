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
    // onChange: (series: DataSeries[]) => void

    /**
     * Live change series while user typing in active series form.
     * Used by consumers for getting latest values for live preview chart.
     * */
    onLiveChange: (liveSeries: DataSeries, isValid: boolean, index: number) => void

    onEditSeriesRequest: (openedCardIndex: number) => void
    onEditSeriesCommit: (seriesIndex: number, editedSeries: DataSeries) => void
    onEditSeriesCancel: (closedCardIndex: number) => void
    onSeriesRemove: (removedSeriesIndex: number) => void
}

/**
 * Renders form series (sub-form) for series (chart lines) creation code insight form.
 * */
export const FormSeries: React.FunctionComponent<FormSeriesProps> = props => {
    const {
        series = [],
        editSeries,
        onEditSeriesRequest,
        onEditSeriesCommit,
        onEditSeriesCancel,
        onSeriesRemove,
        onLiveChange
    } = props

    // In case if we don't have series we have to skip series list ui (components below)
    // and render simple series form component.
    if (series.length === 0) {
        return <FormSeriesInput
            autofocus={false}
            className="card card-body p-3"
            onSubmit={series => onEditSeriesCommit(0, series)}
            onChange={(values, valid) => onLiveChange(values, valid, 0)}
        />
    }

    return (
        <ul className="list-unstyled d-flex flex-column">
            {editSeries.map((line, index) =>
                editSeries[index] ? (
                    <FormSeriesInput
                        key={`${line?.name ?? ''}-${index}`}
                        cancel={true}
                        onSubmit={series => onEditSeriesCommit(index, series)}
                        onCancel={() => onEditSeriesCancel(index)}
                        className={classnames('card card-body p-3', styles.formSeriesItem)}
                        onChange={(values, valid) => onLiveChange(values, valid, index)}
                        {...line}
                    />
                ) : (
                    series[index] &&
                        <SeriesCard
                            key={`${series[index].name}-${index}`}
                            onEdit={() => onEditSeriesRequest(index)}
                            onRemove={() => onSeriesRemove(index)}
                            className={styles.formSeriesItem}
                            {...series[index]}
                        />
                )
            )}

            <button
                type="button"
                onClick={() => onEditSeriesRequest(editSeries.length)}
                className={classnames(styles.formSeriesItem, styles.formSeriesAddButton, 'btn btn-link p-3')}
            >
                + Add another data series
            </button>
        </ul>
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
    const { name, query, stroke: color, className, onEdit, onRemove } = props

    return (
        <li
            aria-label={`Edit button for ${name} data series`}
            className={classnames(styles.formSeriesCard, className, 'card d-flex flex-row p-3')}
        >
            <div className="flex-grow-1 d-flex flex-column align-items-start">

                <div className='d-flex align-items-center mb-1 '>
                    {/* eslint-disable-next-line react/forbid-dom-props */}
                    <div style={{ color }} className={styles.formSeriesCardColor} />
                    <span className="ml-1 font-weight-bold">{name}</span>
                </div>

                <span className="mb-0 text-muted">{query}</span>
            </div>

            <div className='d-flex align-items-center'>

                <button
                    type="button"
                    onClick={onEdit}
                    className='border-0 btn btn-outline-primary'>Edit</button>

                <button
                    type="button"
                    onClick={onRemove}
                    className='border-0 btn btn-outline-danger ml-1'>Remove</button>
            </div>
        </li>
    )
}
