import type { FC, ReactNode } from 'react'

import { noop } from 'rxjs'

import type { FormChangeEvent, SubmissionErrors } from '@sourcegraph/wildcard'

import {
    CreationUiLayout,
    CreationUIForm,
    CreationUIPreview,
    createDefaultEditSeries,
    type EditableDataSeries,
    getSanitizedSeries,
} from '../../../../../components'
import { LineChartLivePreview, type LivePreviewSeries } from '../../LineChartLivePreview'
import type { CreateInsightFormFields } from '../types'

import { type RenderPropertyInputs, SearchInsightCreationForm } from './SearchInsightCreationForm'
import { useInsightCreationForm } from './use-insight-creation-form'

export interface SearchInsightCreationContentProps {
    touched: boolean
    children: (input: RenderPropertyInputs) => ReactNode
    initialValue?: Partial<CreateInsightFormFields>
    dataTestId?: string
    className?: string
    onSubmit: (values: CreateInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    /** Change handlers is called every time when user changed any field within the form. */
    onChange?: (event: FormChangeEvent<CreateInsightFormFields>) => void
}

export const SearchInsightCreationContent: FC<SearchInsightCreationContentProps> = props => {
    const { touched, children, initialValue, className, dataTestId, onSubmit, onChange = noop } = props

    const {
        form: { values, formAPI, handleSubmit },
        title,
        repositories,
        repoQuery,
        repoMode,
        series,
        step,
        stepValue,
    } = useInsightCreationForm({
        touched,
        initialValue,
        onChange,
        onSubmit,
    })

    const handleFormReset = (): void => {
        // TODO [VK] Change useForm API in order to implement form.reset method.
        title.input.onChange('')
        repositories.input.onChange([])
        repoQuery.input.onChange({ query: '' })
        // Focus first element of the form
        repositories.input.ref.current?.focus()
        series.input.onChange([createDefaultEditSeries({ edit: true })])
        stepValue.input.onChange('1')
        step.input.onChange('months')
    }

    // If some fields that needed to run live preview  are invalid
    // we should disable live chart preview
    const allFieldsForPreviewAreValid =
        (repositories.meta.validState === 'VALID' || repoQuery.meta.validState === 'VALID') &&
        (series.meta.validState === 'VALID' || series.input.value.some(series => series.valid)) &&
        stepValue.meta.validState === 'VALID'

    const hasFilledValue =
        values.series?.some(line => line.name !== '' || line.query !== '') ||
        values.repositories.length > 0 ||
        values.repoQuery.query !== '' ||
        values.title !== ''

    return (
        <CreationUiLayout data-testid={dataTestId} className={className}>
            <CreationUIForm
                aria-label="Track changes Insight creation form"
                as={SearchInsightCreationForm}
                handleSubmit={handleSubmit}
                submitErrors={formAPI.submitErrors}
                submitting={formAPI.submitting}
                submitted={formAPI.submitted}
                title={title}
                repositories={repositories}
                repoQuery={repoQuery}
                repoMode={repoMode}
                series={series}
                step={step}
                stepValue={stepValue}
                isFormClearActive={hasFilledValue}
                dashboardReferenceCount={initialValue?.dashboardReferenceCount}
                onFormReset={handleFormReset}
            >
                {children}
            </CreationUIForm>

            <CreationUIPreview
                as={LineChartLivePreview}
                disabled={!allFieldsForPreviewAreValid}
                repositories={repositories.meta.value}
                repoMode={repoMode.meta.value}
                repoQuery={repoQuery.meta.value.query}
                series={seriesToPreview(series.input.value)}
                step={step.meta.value}
                stepValue={stepValue.meta.value}
            />
        </CreationUiLayout>
    )
}

function seriesToPreview(currentSeries: EditableDataSeries[]): LivePreviewSeries[] {
    const validSeries = currentSeries.filter(series => series.valid)
    return getSanitizedSeries(validSeries).map(series => ({
        query: series.query,
        stroke: series.stroke ? series.stroke : '',
        label: series.name,
        generatedFromCaptureGroup: false,
    }))
}
