import classnames from 'classnames'
import React, { forwardRef, ReactElement, useCallback, useImperativeHandle, useRef, useState } from 'react'

import { DataSeries } from '../../types'
import { FormSeriesInput, FormSeriesInputAPI } from '../form-series-input/FormSeriesInput'

import styles from './FormSeries.module.scss'

export interface FormSeriesProps {
    /** Name of form series input (sub-form) */
    name: string
    /** Controlled value (series - chart lines) for series input component. */
    series?: DataSeries[]
    /** Change handler. */
    onChange: (series: DataSeries[]) => void
}

/**
 * Public API - mimic the standard native input API.
 * Consumers of this form (FormSeries) may have to able to focus
 * some input of this form. For it we have to provider (expose) some api.
 * */
export interface FormSeriesReferenceAPI {
    /** Mimic-value of native input name. */
    name: string
    /** Mimic-function of native input focus. */
    focus: () => void
}

/** Renders form series (sub-form) for series (chart lines) creation code insight form. */
export const FormSeries = forwardRef<FormSeriesReferenceAPI, FormSeriesProps>((props, reference) => {
    const { name, series = [], onChange } = props

    // We can have more than one series. In case if user clicked on existing series
    // he should see edit form for series. To track which series is active now
    // we use array of theses series indexes.
    const [editSeriesIndexes, setEditSeriesIndex] = useState<number[]>([])
    const [newSeriesEdit, setNewSeriesEdit] = useState(false)

    const seriesInputReference = useRef<FormSeriesInputAPI>(null)

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

    // In some cases consumers of this component may want to call focus or get name field
    // in a way that would be a native html element.
    useImperativeHandle(reference, () => ({
        name,
        focus: () => seriesInputReference.current?.focus(),
    }))

    // In case if we don't have series we have to skip series list ui (components below)
    // and render simple series form component.
    if (series.length === 0) {
        return (
            <FormSeriesInput
                innerRef={seriesInputReference}
                className={styles.formSeriesInput}
                onSubmit={handleSubmitNewSeries}
            />
        )
    }

    return (
        <div className='d-flex flex-column'>
            {series.map((line, index) =>
                editSeriesIndexes.includes(index) ? (
                    <FormSeriesInput
                        key={`${line.name}-${index}`}
                        autofocus={true}
                        cancel={true}
                        /* eslint-disable-next-line react/jsx-no-bind */
                        onSubmit={series => handleEditSeriesCommit(index, series)}
                        /* eslint-disable-next-line react/jsx-no-bind */
                        onCancel={() => handleCancelEditSeries(index)}
                        className={classnames(styles.formSeriesInput, styles.formSeriesItem)}
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
                    autofocus={true}
                    innerRef={seriesInputReference}
                    cancel={true}
                    onSubmit={handleSubmitNewSeries}
                    onCancel={handleCancelNewSeries}
                    className={classnames(styles.formSeriesInput, styles.formSeriesItem)}
                />
            )}

            {!newSeriesEdit && (
                <button
                    type="button"
                    onClick={handleAddSeries}
                    className={classnames(styles.formSeriesItem, styles.formSeriesAddButton, 'btn btn-link')}
                >
                    + Add another data series
                </button>
            )}
        </div>
    )
})

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
            className={classnames(styles.formSeriesCard, className, 'd-flex p-3 btn btn-outline-secondary')}>

            <div className="flex-grow-1 d-flex flex-column align-items-start">
                <span className='mb-1 font-weight-bold'>{name}</span>
                <span className='mb-0 text-muted'>{query}</span>
            </div>

            {/* eslint-disable-next-line react/forbid-dom-props */}
            <div style={{ color }} className={styles.formSeriesCardColor} />
        </button>
    )
}
