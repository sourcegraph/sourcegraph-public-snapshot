import classnames from 'classnames'
import React, { ReactElement, useCallback, useState } from 'react'

import { DataSeries } from '../../types'
import { FormSeriesInput } from '../form-series-input/FormSeriesInput'

import styles from './FormSeries.module.scss'

export interface FormSeriesProps {
    /** Controlled value (series - chart lines) for series input component. */
    series?: DataSeries[]
    /** Change handler. */
    onChange: (series: DataSeries[]) => void
}

/** Renders form series (sub-form) for series (chart lines) creation code insight form. */
export const FormSeries: React.FunctionComponent<FormSeriesProps> = props => {
    const { series = [], onChange } = props

    // We can have more than one series. In case if user clicked on existing series
    // he should see edit form for series. To track which series is active now
    // we use array of theses series indexes.
    const [editSeriesIndexes, setEditSeriesIndex] = useState<number[]>([])
    const [newSeriesEdit, setNewSeriesEdit] = useState(false)

    // Add empty series form component that user be able to fill this form and create
    // another chart series.
    const handleAddSeries = useCallback(() => {
        setNewSeriesEdit(true)
    }, [setNewSeriesEdit])

    // Close new series form component in case if user clicked cancel button.
    const handleCancelNewSeries = useCallback(() => setNewSeriesEdit(false), [setNewSeriesEdit])

    // Handle submit of creation of new chart series.
    const handleSubmitNewSeries = useCallback(
        (newSeries: DataSeries) => {
            // Close series input in case if we add another series
            if (newSeriesEdit) {
                setNewSeriesEdit(false)
            }

            onChange([...series, newSeries])
        },
        [series, newSeriesEdit, setNewSeriesEdit, onChange]
    )

    const handleEditSeriesCommit = (index: number, editedSeries: DataSeries): void => {
        const newSeries = [...series]

        newSeries[index] = editedSeries
        setEditSeriesIndex(indexes => indexes.filter(currentIndex => currentIndex !== index))
        onChange(newSeries)
    }

    const handleRequestToEdit = (index: number): void => {
        setEditSeriesIndex([...editSeriesIndexes, index])
    }

    const handleCancelEditSeries = (index: number): void => {
        setEditSeriesIndex(indexes => indexes.filter(currentIndex => currentIndex !== index))
    }

    // In case if we don't have series we have to skip series list ui (components below)
    // and render simple series form component.
    if (series.length === 0) {
        return <FormSeriesInput autofocus={false} className="card card-body p-3" onSubmit={handleSubmitNewSeries} />
    }

    return (
        <div role="list" className="d-flex flex-column">
            {series.map((line, index) =>
                editSeriesIndexes.includes(index) ? (
                    <FormSeriesInput
                        key={`${line.name}-${index}`}
                        cancel={true}
                        /* eslint-disable-next-line react/jsx-no-bind */
                        onSubmit={series => handleEditSeriesCommit(index, series)}
                        /* eslint-disable-next-line react/jsx-no-bind */
                        onCancel={() => handleCancelEditSeries(index)}
                        className={classnames('card card-body p-3', styles.formSeriesItem)}
                        {...line}
                    />
                ) : (
                    <SeriesCard
                        key={`${line.name}-${index}`}
                        /* eslint-disable-next-line react/jsx-no-bind */
                        onEdit={() => handleRequestToEdit(index)}
                        className={styles.formSeriesItem}
                        {...line}
                    />
                )
            )}

            {newSeriesEdit && (
                <FormSeriesInput
                    cancel={true}
                    onSubmit={handleSubmitNewSeries}
                    onCancel={handleCancelNewSeries}
                    className={classnames('card card-body p-3', styles.formSeriesItem)}
                />
            )}

            {!newSeriesEdit && (
                <button
                    type="button"
                    onClick={handleAddSeries}
                    className={classnames(styles.formSeriesItem, styles.formSeriesAddButton, 'btn btn-link p-3')}
                >
                    + Add another data series
                </button>
            )}
        </div>
    )
}

interface SeriesCardProps {
    /** Name of series. */
    name: string
    /** Query value of series. */
    query: string
    /** Color value of series. */
    color: string
    /** Custom class name for root button element. */
    className?: string
    /** Edit handler. */
    onEdit?: () => void
    /** Remove handler. */
    onRemove?: () => void
}

/** Renders series card component, visual list item of series (name, color, query) */
function SeriesCard(props: SeriesCardProps): ReactElement {
    const { name, query, color, className, onEdit } = props

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
