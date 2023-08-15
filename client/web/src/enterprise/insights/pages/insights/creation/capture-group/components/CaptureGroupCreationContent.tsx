import type { FC, ReactNode } from 'react'

import { noop } from 'lodash'

import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import { useField, type FormChangeEvent, type SubmissionErrors, useForm } from '@sourcegraph/wildcard'

import { CreationUiLayout, CreationUIForm, CreationUIPreview, useRepoFields } from '../../../../../components'
import { LineChartLivePreview } from '../../LineChartLivePreview'
import type { CaptureGroupFormFields } from '../types'

import { CaptureGroupCreationForm, type RenderPropertyInputs } from './CaptureGoupCreationForm'
import { QUERY_VALIDATORS, STEP_VALIDATORS, TITLE_VALIDATORS } from './validators'

const INITIAL_VALUES: CaptureGroupFormFields = {
    repositories: [],
    groupSearchQuery: '',
    title: '',
    step: 'months',
    stepValue: '2',
    repoMode: 'search-query',
    repoQuery: { query: '' },
    dashboardReferenceCount: 0,
}

interface CaptureGroupCreationContentProps {
    initialValues?: Partial<CaptureGroupFormFields>
    touched: boolean
    className?: string
    children: (inputs: RenderPropertyInputs) => ReactNode
    onSubmit: (values: CaptureGroupFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onChange?: (event: FormChangeEvent<CaptureGroupFormFields>) => void
    onCancel: () => void
}

export const CaptureGroupCreationContent: FC<CaptureGroupCreationContentProps> = props => {
    const { touched, initialValues = {}, className, children, onSubmit, onChange = noop } = props

    const repoFieldVariation = useExperimentalFeatures(features => features.codeInsightsRepoUI)
    const isSearchQueryORUrlsList = repoFieldVariation === 'search-query-or-strict-list'

    // Enforce "search-query" initial value if we're in the single search query UI mode
    const fixedInitialValues = isSearchQueryORUrlsList
        ? { ...INITIAL_VALUES, ...initialValues }
        : { ...INITIAL_VALUES, ...initialValues, repoMode: 'search-query' as const }

    const form = useForm<CaptureGroupFormFields>({
        initialValues: fixedInitialValues,
        touched,
        onSubmit,
        onChange,
    })

    const { repoMode, repoQuery, repositories } = useRepoFields({ formApi: form.formAPI })

    const title = useField({
        name: 'title',
        formApi: form.formAPI,
        validators: { sync: TITLE_VALIDATORS },
    })

    const query = useField({
        name: 'groupSearchQuery',
        formApi: form.formAPI,
        validators: { sync: QUERY_VALIDATORS },
    })

    const step = useField({
        name: 'step',
        formApi: form.formAPI,
    })

    const stepValue = useField({
        name: 'stepValue',
        formApi: form.formAPI,
        validators: { sync: STEP_VALIDATORS },
    })

    const handleFormReset = (): void => {
        title.input.onChange('')
        repoQuery.input.onChange({ query: '' })
        repositories.input.onChange([])
        query.input.onChange('')
        step.input.onChange('months')
        stepValue.input.onChange('1')

        // Focus first element of the form
        repositories.input.ref.current?.focus()
    }

    const hasFilledValue =
        form.values.title !== '' ||
        form.values.repositories.length > 0 ||
        form.values.repoQuery.query !== '' ||
        form.values.groupSearchQuery !== ''

    const areAllFieldsForPreviewValid =
        (repositories.meta.validState === 'VALID' || repoQuery.meta.validState === 'VALID') &&
        stepValue.meta.validState === 'VALID' &&
        query.meta.validState === 'VALID'

    return (
        <CreationUiLayout className={className}>
            <CreationUIForm
                aria-label="Detect and track Insight creation form"
                as={CaptureGroupCreationForm}
                form={form}
                title={title}
                repoMode={repoMode}
                repoQuery={repoQuery}
                repositories={repositories}
                step={step}
                stepValue={stepValue}
                query={query}
                isFormClearActive={hasFilledValue}
                dashboardReferenceCount={initialValues.dashboardReferenceCount}
                onFormReset={handleFormReset}
            >
                {children}
            </CreationUIForm>

            <CreationUIPreview
                as={LineChartLivePreview}
                disabled={!areAllFieldsForPreviewValid}
                repositories={repositories.meta.value}
                repoMode={repoMode.meta.value}
                repoQuery={repoQuery.meta.value.query}
                series={captureGroupPreviewSeries(query.meta.value)}
                step={step.meta.value}
                stepValue={stepValue.meta.value}
            />
        </CreationUiLayout>
    )
}

function captureGroupPreviewSeries(query: string): any[] {
    return [
        {
            generatedFromCaptureGroup: true,
            label: '',
            query,
            stroke: '',
        },
    ]
}
