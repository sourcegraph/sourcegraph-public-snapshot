import classnames from 'classnames'
import React from 'react'
import { noop } from 'rxjs'

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { useField } from '../../../../../components/form/hooks/useField'
import { FormChangeEvent, SubmissionErrors, useForm } from '../../../../../components/form/hooks/useForm'
import { useTitleValidator } from '../../../../../components/form/hooks/useTitleValidator'
import { Organization } from '../../../../../components/visibility-picker/VisibilityPicker'
import { InsightTypePrefix } from '../../../../../core/types'
import { CreateInsightFormFields } from '../../types'
import { SearchInsightLivePreview } from '../live-preview-chart/SearchInsightLivePreview'
import { SearchInsightCreationForm } from '../search-insight-creation-form/SearchInsightCreationForm'

import { useEditableSeries, createDefaultEditSeries } from './hooks/use-editable-series'
import styles from './SearchInsightCreationContent.module.scss'
import {
    repositoriesExistValidator,
    repositoriesFieldValidator,
    requiredStepValueField,
    seriesRequired,
} from './validators'

const INITIAL_VALUES: CreateInsightFormFields = {
    visibility: 'personal',
    // If user opens creation form to create insight
    // we want to show the series form as soon as possible without
    // force user to click 'add another series' button
    series: [createDefaultEditSeries({ edit: true })],
    step: 'months',
    stepValue: '2',
    title: '',
    repositories: '',
}

export interface SearchInsightCreationContentProps {
    /** This component might be used in edit or creation insight case. */
    mode?: 'creation' | 'edit'
    /** Final settings cascade. Used for title field validation. */
    settings?: Settings | null
    /** List of all user organizations */
    organizations?: Organization[]
    /** Initial value for all form fields. */
    initialValue?: CreateInsightFormFields
    /** Custom class name for root form element. */
    className?: string
    /** Test id for the root content element (form element). */
    dataTestId?: string
    /** Submit handler for form element. */
    onSubmit: (values: CreateInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    /** Cancel handler. */
    onCancel?: () => void
    /** Change handlers is called every time when user changed any field within the form. */
    onChange?: (event: FormChangeEvent<CreateInsightFormFields>) => void
}

export const SearchInsightCreationContent: React.FunctionComponent<SearchInsightCreationContentProps> = props => {
    const {
        mode = 'creation',
        organizations = [],
        settings,
        initialValue = INITIAL_VALUES,
        className,
        dataTestId,
        onSubmit,
        onCancel = noop,
        onChange = noop,
    } = props

    const isEditMode = mode === 'edit'

    const { values, formAPI, ref, handleSubmit } = useForm<CreateInsightFormFields>({
        initialValues: initialValue,
        onSubmit,
        onChange,
        touched: isEditMode,
    })

    // We can't have two or more insights with the same name, since we rely on name as on id of insights.
    const titleValidator = useTitleValidator({ settings, insightType: InsightTypePrefix.search })

    const title = useField('title', formAPI, { sync: titleValidator })
    const repositories = useField('repositories', formAPI, {
        sync: repositoriesFieldValidator,
        async: repositoriesExistValidator,
    })
    const visibility = useField('visibility', formAPI)

    const series = useField('series', formAPI, { sync: seriesRequired })
    const step = useField('step', formAPI)
    const stepValue = useField('stepValue', formAPI, { sync: requiredStepValueField })

    const { editSeries, listen, editRequest, editCommit, cancelEdit, deleteSeries } = useEditableSeries({
        series,
    })

    const handleFormReset = (): void => {
        // TODO [VK] Change useForm API in order to implement form.reset method.
        title.input.onChange('')
        repositories.input.onChange('')
        // Focus first element of the form
        repositories.input.ref.current?.focus()
        visibility.input.onChange('personal')
        series.input.onChange([createDefaultEditSeries({ edit: true })])
        stepValue.input.onChange('2')
        step.input.onChange('months')
    }

    const validEditSeries = editSeries.filter(series => series.valid)

    // If some fields that needed to run live preview  are invalid
    // we should disabled live chart preview
    const allFieldsForPreviewAreValid =
        (repositories.meta.validState === 'VALID' || repositories.meta.validState === 'CHECKING') &&
        (series.meta.validState === 'VALID' || validEditSeries.length) &&
        stepValue.meta.validState === 'VALID'

    const hasFilledValue =
        values.series?.some(line => line.name !== '' || line.query !== '') ||
        values.repositories !== '' ||
        values.title !== ''

    return (
        <div data-testid={dataTestId} className={classnames(styles.content, className)}>
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
                visibility={visibility}
                organizations={organizations}
                series={series}
                step={step}
                stepValue={stepValue}
                isFormClearActive={hasFilledValue}
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
                series={editSeries}
                step={step.meta.value}
                stepValue={stepValue.meta.value}
                className={styles.contentLivePreview}
            />
        </div>
    )
}
