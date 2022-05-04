import React from 'react'

import classNames from 'classnames'

import { Button } from '@sourcegraph/wildcard'

import { LimitedAccessLabel } from '../../../../../../components/limited-access-label/LimitedAccessLabel'
import { useUiFeatures } from '../../../../../../hooks/use-ui-features'
import { EditableDataSeries } from '../../types'
import { FormSeriesInput } from '../form-series-input/FormSeriesInput'

import { SeriesCard } from './components/series-card/SeriesCard'

import styles from './FormSeries.module.scss'

export interface FormSeriesProps {
    /**
     * Show all validation error for all forms and fields within the series forms.
     */
    showValidationErrorsOnMount: boolean

    /**
     * Controlled value (series - chart lines) for series input component.
     */
    series?: EditableDataSeries[]

    /**
     * Code Insight repositories field string value - repo1, repo2, ...
     */
    repositories: string

    /**
     * Live change series handler while user typing in active series form.
     * Used by consumers to get latest values from series inputs and pass
     * them tp live preview chart.
     */
    onLiveChange: (liveSeries: EditableDataSeries, isValid: boolean, index: number) => void

    /**
     * Handler that runs every time user clicked edit on particular
     * series card.
     */
    onEditSeriesRequest: (seriesId?: string) => void

    /**
     * Handler that runs every time use clicked commit (done) in
     * series edit form.
     */
    onEditSeriesCommit: (editedSeries: EditableDataSeries) => void

    /**
     * Handler that runs every time use canceled (click cancel) in
     * series edit form.
     */
    onEditSeriesCancel: (seriesId: string) => void

    /**
     * Handler that runs every time use removed (click remove) in
     * series card.
     */
    onSeriesRemove: (seriesId: string) => void
}

/**
 * Renders form series (sub-form) for series (chart lines) creation code insight form.
 */
export const FormSeries: React.FunctionComponent<React.PropsWithChildren<FormSeriesProps>> = props => {
    const {
        series = [],
        showValidationErrorsOnMount,
        repositories,
        onEditSeriesRequest,
        onEditSeriesCommit,
        onEditSeriesCancel,
        onSeriesRemove,
        onLiveChange,
    } = props

    const { licensed } = useUiFeatures()

    return (
        <ul data-testid="form-series" className="list-unstyled d-flex flex-column">
            {series.map((line, index) =>
                line.edit ? (
                    <FormSeriesInput
                        key={line.id}
                        series={line}
                        showValidationErrorsOnMount={showValidationErrorsOnMount}
                        index={index + 1}
                        cancel={series.length > 1}
                        autofocus={series.length > 1}
                        repositories={repositories}
                        onSubmit={onEditSeriesCommit}
                        onCancel={() => onEditSeriesCancel(line.id)}
                        className={classNames('p-3', styles.formSeriesItem)}
                        onChange={(seriesValues, valid) => onLiveChange({ ...line, ...seriesValues }, valid, index)}
                    />
                ) : (
                    line && (
                        <SeriesCard
                            key={line.id}
                            disabled={index >= 10}
                            onEdit={() => onEditSeriesRequest(line.id)}
                            onRemove={() => onSeriesRemove(line.id)}
                            className={styles.formSeriesItem}
                            {...line}
                        />
                    )
                )
            )}

            {!licensed && (
                <LimitedAccessLabel message="Unlock Code Insights for unlimited data series" className="mx-auto my-3" />
            )}

            <Button
                data-testid="add-series-button"
                type="button"
                onClick={() => onEditSeriesRequest()}
                variant="link"
                disabled={!licensed ? series.length >= 10 : false}
                className={classNames(styles.formSeriesItem, styles.formSeriesAddButton, 'p-3')}
            >
                + Add another data series
            </Button>
        </ul>
    )
}
