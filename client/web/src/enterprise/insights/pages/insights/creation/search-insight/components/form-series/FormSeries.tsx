import React from 'react'

import classNames from 'classnames'

import { Button } from '@sourcegraph/wildcard'

import { LimitedAccessLabel, useFieldAPI } from '../../../../../../components'
import { useUiFeatures } from '../../../../../../hooks'
import { EditableDataSeries } from '../../types'
import { FormSeriesInput } from '../form-series-input/FormSeriesInput'

import { SeriesCard } from './components/series-card/SeriesCard'
import { useEditableSeries } from './hooks/use-editable-series'

import styles from './FormSeries.module.scss'

export interface FormSeriesProps {
    seriesField: useFieldAPI<EditableDataSeries[]>
    repositories: string
    showValidationErrorsOnMount: boolean
}

export const FormSeries: React.FunctionComponent<React.PropsWithChildren<FormSeriesProps>> = props => {
    const { seriesField, showValidationErrorsOnMount, repositories } = props

    const { licensed } = useUiFeatures()
    const { series, changeSeries, editRequest, editCommit, cancelEdit, deleteSeries } = useEditableSeries(seriesField)

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
                        onSubmit={editCommit}
                        onCancel={() => cancelEdit(line.id)}
                        className={classNames('p-3', styles.formSeriesItem)}
                        onChange={(seriesValues, valid) => changeSeries(seriesValues, valid, index)}
                    />
                ) : (
                    line && (
                        <SeriesCard
                            key={line.id}
                            disabled={index >= 10}
                            onEdit={() => editRequest(line.id)}
                            onRemove={() => deleteSeries(line.id)}
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
                onClick={() => editRequest()}
                variant="link"
                disabled={!licensed ? series.length >= 10 : false}
                className={classNames(styles.formSeriesItem, styles.formSeriesAddButton, 'p-3')}
            >
                + Add another data series
            </Button>
        </ul>
    )
}
