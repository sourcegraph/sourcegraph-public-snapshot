import { FC, ReactNode } from 'react'

import { noop } from 'lodash'

import {
    CreationUiLayout,
    CreationUIForm,
    CreationUIPreview,
    useField,
    FormChangeEvent,
    SubmissionErrors,
    useForm,
    insightRepositoriesValidator,
    insightRepositoriesAsyncValidator,
} from '../../../../../components'
import { LineChartLivePreview } from '../../LineChartLivePreview'
import { CaptureGroupFormFields } from '../types'

import { CaptureGroupCreationForm, RenderPropertyInputs } from './CaptureGoupCreationForm'
import { QUERY_VALIDATORS, STEP_VALIDATORS, TITLE_VALIDATORS } from './validators'

const INITIAL_VALUES: CaptureGroupFormFields = {
    repositories: '',
    groupSearchQuery: '',
    title: '',
    step: 'months',
    stepValue: '2',
    allRepos: false,
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

    const form = useForm<CaptureGroupFormFields>({
        initialValues: { ...INITIAL_VALUES, ...initialValues },
        touched,
        onSubmit,
        onChange,
    })

    const title = useField({
        name: 'title',
        formApi: form.formAPI,
        validators: { sync: TITLE_VALIDATORS },
    })

    const allReposMode = useField({
        name: 'allRepos',
        formApi: form.formAPI,
        onChange: (checked: boolean) => {
            // Reset form values in case if All repos mode was activated
            if (checked) {
                repositories.input.onChange('')
            }
        },
    })

    const repositories = useField({
        name: 'repositories',
        formApi: form.formAPI,
        validators: {
            // Turn off any validations for the repositories' field in we are in all repos mode
            sync: !allReposMode.input.value ? insightRepositoriesValidator : undefined,
            async: !allReposMode.input.value ? insightRepositoriesAsyncValidator : undefined,
        },
        disabled: allReposMode.input.value,
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
        repositories.input.onChange('')
        query.input.onChange('')
        step.input.onChange('months')
        stepValue.input.onChange('1')

        // Focus first element of the form
        repositories.input.ref.current?.focus()
    }

    const hasFilledValue =
        form.values.title !== '' || form.values.repositories !== '' || form.values.groupSearchQuery !== ''

    const areAllFieldsForPreviewValid =
        repositories.meta.validState === 'VALID' &&
        stepValue.meta.validState === 'VALID' &&
        query.meta.validState === 'VALID' &&
        // For all repos mode we are not able to show the live preview chart
        !allReposMode.input.value

    return (
        <CreationUiLayout className={className}>
            <CreationUIForm
                aria-label="Detect and track Insight creation form"
                as={CaptureGroupCreationForm}
                form={form}
                title={title}
                repositories={repositories}
                step={step}
                stepValue={stepValue}
                query={query}
                isFormClearActive={hasFilledValue}
                allReposMode={allReposMode}
                dashboardReferenceCount={initialValues.dashboardReferenceCount}
                onFormReset={handleFormReset}
            >
                {children}
            </CreationUIForm>

            <CreationUIPreview
                as={LineChartLivePreview}
                disabled={!areAllFieldsForPreviewValid}
                isAllReposMode={allReposMode.input.value}
                repositories={repositories.meta.value}
                series={captureGroupPreviewSeries(query.meta.value)}
                step={step.meta.value}
                stepValue={stepValue.meta.value}
            />
        </CreationUiLayout>
    )
}

function captureGroupPreviewSeries(query: string): any {
    return [
        {
            generatedFromCaptureGroup: true,
            label: '',
            query,
            stroke: '',
        },
    ]
}
