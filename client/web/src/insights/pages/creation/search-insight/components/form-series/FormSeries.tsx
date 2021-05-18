import classnames from 'classnames'
import React from 'react'

import { DataSeries } from '../../../../../core/backend/types'
import { FormSeriesInput } from '../form-series-input/FormSeriesInput'

import { SeriesCard } from './components/series-card/SeriesCard'
import styles from './FormSeries.module.scss'

export interface FormSeriesProps {
    /**
     * Controlled value (series - chart lines) for series input component.
     * */
    series?: DataSeries[]

    /**
     * Edit series used below for rendering edit series form. Element of
     * this array has undefined value when there are no series to edit
     * and has DataSeries value when user activated edit for some series.
     * */
    editSeries: (DataSeries | undefined)[]

    /**
     * Live change series handler while user typing in active series form.
     * Used by consumers to get latest values from series inputs and pass
     * them tp live preview chart.
     * */
    onLiveChange: (liveSeries: DataSeries, isValid: boolean, index: number) => void

    /**
     * Handler that runs every time user clicked edit on particular
     * series card.
     * */
    onEditSeriesRequest: (editSeriesIndex: number) => void

    /**
     * Handler that runs every time use clicked commit (done) in
     * series edit form.
     * */
    onEditSeriesCommit: (seriesIndex: number, editedSeries: DataSeries) => void

    /**
     * Handler that runs every time use canceled (click cancel) in
     * series edit form.
     * */
    onEditSeriesCancel: (closedCardIndex: number) => void

    /**
     * Handler that runs every time use removed (click remove) in
     * series card.
     * */
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
        onLiveChange,
    } = props

    // In case if we don't have series we have to skip series list ui (components below)
    // and render simple series form component.
    if (series.length === 0) {
        return (
            <FormSeriesInput
                index={1}
                autofocus={false}
                className="card card-body p-3"
                onSubmit={series => onEditSeriesCommit(0, series)}
                onChange={(values, valid) => onLiveChange(values, valid, 0)}
            />
        )
    }

    return (
        <ul className="list-unstyled d-flex flex-column">
            {editSeries.map((line, index) =>
                editSeries[index] ? (
                    <FormSeriesInput
                        key={`${line?.name ?? ''}-${index}`}
                        index={index + 1}
                        cancel={true}
                        onSubmit={series => onEditSeriesCommit(index, series)}
                        onCancel={() => onEditSeriesCancel(index)}
                        className={classnames('card card-body p-3', styles.formSeriesItem)}
                        onChange={(values, valid) => onLiveChange(values, valid, index)}
                        {...line}
                    />
                ) : (
                    series[index] && (
                        <SeriesCard
                            key={`${series[index].name}-${index}`}
                            onEdit={() => onEditSeriesRequest(index)}
                            onRemove={() => onSeriesRemove(index)}
                            className={styles.formSeriesItem}
                            {...series[index]}
                        />
                    )
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
