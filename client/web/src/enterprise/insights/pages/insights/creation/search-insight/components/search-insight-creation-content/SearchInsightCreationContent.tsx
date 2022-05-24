import React from 'react'

import classNames from 'classnames'
import { noop } from 'rxjs'

import { styles } from '../../../../../../components/creation-ui-kit'
import { FormChangeEvent, SubmissionErrors } from '../../../../../../components/form/hooks/useForm'
import { Insight } from '../../../../../../core'
import { CreateInsightFormFields } from '../../types'
import { SearchInsightCreationForm } from '../SearchInsightCreationForm'
import { SearchInsightLivePreview } from '../SearchInsightLivePreview'

import { useEditableSeries, createDefaultEditSeries } from './hooks/use-editable-series'
import { useInsightCreationForm } from './hooks/use-insight-creation-form/use-insight-creation-form'

export interface SearchInsightCreationContentProps {
    /** This component might be used in edit or creation insight case. */
    mode?: 'creation' | 'edit'

    initialValue?: Partial<CreateInsightFormFields>
    className?: string
    dataTestId?: string
    insight?: Insight

    onSubmit: (values: CreateInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel?: () => void

    /** Change handlers is called every time when user changed any field within the form. */
    onChange?: (event: FormChangeEvent<CreateInsightFormFields>) => void
}

export const SearchInsightCreationContent: React.FunctionComponent<
    React.PropsWithChildren<SearchInsightCreationContentProps>
> = props => {
    const {
        mode = 'creation',
        initialValue,
        className,
        dataTestId,
        insight,
        onSubmit,
        onCancel = noop,
        onChange = noop,
    } = props

    const {
        form: { values, formAPI, ref, handleSubmit },
        title,
        repositories,
        series,
        step,
        stepValue,
        allReposMode,
    } = useInsightCreationForm({
        mode,
        initialValue,
        onChange,
        onSubmit,
    })

    const { editSeries, listen, editRequest, editCommit, cancelEdit, deleteSeries } = useEditableSeries({ series })

    const handleFormReset = (): void => {
        // TODO [VK] Change useForm API in order to implement form.reset method.
        title.input.onChange('')
        repositories.input.onChange('')
        // Focus first element of the form
        repositories.input.ref.current?.focus()
        series.input.onChange([createDefaultEditSeries({ edit: true })])
        stepValue.input.onChange('1')
        step.input.onChange('months')
    }

    // If some fields that needed to run live preview  are invalid
    // we should disable live chart preview
    const allFieldsForPreviewAreValid =
        repositories.meta.validState === 'VALID' &&
        (series.meta.validState === 'VALID' || editSeries.some(series => series.valid)) &&
        stepValue.meta.validState === 'VALID' &&
        // For the "all repositories" mode we are not able to show the live preview chart
        !allReposMode.input.value

    const hasFilledValue =
        values.series?.some(line => line.name !== '' || line.query !== '') ||
        values.repositories !== '' ||
        values.title !== ''

    return (
        <div data-testid={dataTestId} className={classNames(styles.content, className)}>
            <SearchInsightCreationForm
                mode={mode}
                className={styles.contentForm}
                innerRef={ref}
                handleSubmit={handleSubmit}
                submitErrors={formAPI.submitErrors}
                submitting={formAPI.submitting}
                submitted={formAPI.submitted}
                title={title}
                repositories={repositories}
                allReposMode={allReposMode}
                series={series}
                step={step}
                stepValue={stepValue}
                isFormClearActive={hasFilledValue}
                dashboardReferenceCount={initialValue?.dashboardReferenceCount}
                insight={insight}
                onSeriesLiveChange={listen}
                onCancel={onCancel}
                onEditSeriesRequest={editRequest}
                onEditSeriesCancel={cancelEdit}
                onEditSeriesCommit={editCommit}
                onSeriesRemove={deleteSeries}
                onFormReset={handleFormReset}
            />

            <SearchInsightLivePreview
                disabled={!allFieldsForPreviewAreValid}
                repositories={repositories.meta.value}
                isAllReposMode={allReposMode.input.value}
                series={editSeries}
                step={step.meta.value}
                stepValue={stepValue.meta.value}
                className={styles.contentLivePreview}
            />
        </div>
    )
}
