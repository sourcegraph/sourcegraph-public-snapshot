import type { FC, ReactNode } from 'react'

import classNames from 'classnames'

import { Button, type useFieldAPI } from '@sourcegraph/wildcard'

import { useUiFeatures } from '../../../hooks'
import { LimitedAccessLabel } from '../../index'

import { FormSeriesInput } from './components/form-series-input/FormSeriesInput'
import { SeriesCard } from './components/series-card/SeriesCard'
import type { EditableDataSeries } from './types'
import { useEditableSeries } from './use-editable-series'

import styles from './FormSeries.module.scss'

export interface FormSeriesProps {
    seriesField: useFieldAPI<EditableDataSeries[]>
    repositories: string[]
    repoQuery: string | null
    showValidationErrorsOnMount: boolean

    /**
     * This field is only needed for specifying a special compute-specific
     * query field description when this component is used on the compute-powered insight.
     * This prop should be removed when we will have a better form series management
     * solution, see https://github.com/sourcegraph/sourcegraph/issues/38236
     */
    queryFieldDescription?: ReactNode

    /**
     * For the compute-powered insight we do not support multi series, in order to enforce it
     * we need to hide this functionality by hiding add new series button.
     * More context in this issue https://github.com/sourcegraph/sourcegraph/issues/38832
     */
    hasAddNewSeriesButton?: boolean
}

export const FormSeries: FC<FormSeriesProps> = props => {
    const {
        seriesField,
        showValidationErrorsOnMount,
        repoQuery,
        repositories,
        queryFieldDescription,
        hasAddNewSeriesButton = true,
    } = props

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
                        autofocus={line.autofocus}
                        repoQuery={repoQuery}
                        repositories={repositories}
                        queryFieldDescription={queryFieldDescription}
                        className={classNames('p-3', styles.formSeriesItem)}
                        onSubmit={editCommit}
                        onCancel={() => cancelEdit(line.id)}
                        onChange={(seriesValues, valid) => changeSeries(seriesValues, valid, index)}
                    />
                ) : (
                    line && (
                        <SeriesCard
                            key={line.id}
                            disabled={!licensed ? index >= 10 : false}
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

            {hasAddNewSeriesButton && (
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
            )}
        </ul>
    )
}
