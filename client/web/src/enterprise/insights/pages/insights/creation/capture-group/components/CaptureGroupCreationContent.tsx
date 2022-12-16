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
    useRepoFields,
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
    repoMode: 'urls-list',
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

    const form = useForm<CaptureGroupFormFields>({
        initialValues: { ...INITIAL_VALUES, ...initialValues },
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
        repositories.input.onChange('')
        query.input.onChange('')
        step.input.onChange('months')
        stepValue.input.onChange('1')

        // Focus first element of the form
        repositories.input.ref.current?.focus()
    }

    const hasFilledValue =
        form.values.title !== '' ||
        form.values.repositories !== '' ||
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
