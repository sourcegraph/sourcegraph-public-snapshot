import type { FC, FormEventHandler, ReactNode, FormHTMLAttributes } from 'react'

import { Input, FormGroup, getDefaultInputProps, type useFieldAPI, type SubmissionErrors } from '@sourcegraph/wildcard'

import {
    FormSeries,
    CodeInsightDashboardsVisibility,
    CodeInsightTimeStepPicker,
    RepoSettingSection,
} from '../../../../../components'
import { useUiFeatures } from '../../../../../hooks'
import type { CreateInsightFormFields } from '../types'

interface CreationSearchInsightFormProps extends Omit<FormHTMLAttributes<HTMLFormElement>, 'title' | 'children'> {
    handleSubmit: FormEventHandler
    submitErrors: SubmissionErrors
    submitting: boolean
    submitted: boolean
    isFormClearActive: boolean
    dashboardReferenceCount?: number

    title: useFieldAPI<CreateInsightFormFields['title']>
    repositories: useFieldAPI<CreateInsightFormFields['repositories']>
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
    repoMode: useFieldAPI<CreateInsightFormFields['repoMode']>

    series: useFieldAPI<CreateInsightFormFields['series']>
    step: useFieldAPI<CreateInsightFormFields['step']>
    stepValue: useFieldAPI<CreateInsightFormFields['stepValue']>

    children: (inputs: RenderPropertyInputs) => ReactNode
    onFormReset: () => void
}

export interface RenderPropertyInputs {
    submitting: boolean
    submitErrors: SubmissionErrors
    isFormClearActive: boolean
}

/**
 * Displays creation code insight form (title, visibility, series, etc.)
 * UI layer only, all controlled data should be managed by consumer of this component.
 */
export const SearchInsightCreationForm: FC<CreationSearchInsightFormProps> = props => {
    const {
        handleSubmit,
        submitErrors,
        submitting,
        submitted,
        title,
        repositories,
        repoQuery,
        repoMode,
        series,
        stepValue,
        step,
        isFormClearActive,
        dashboardReferenceCount,
        children,
        onFormReset,
        ...attributes
    } = props

    const { licensed } = useUiFeatures()

    return (
        // eslint-disable-next-line react/forbid-elements
        <form {...attributes} noValidate={true} onSubmit={handleSubmit} onReset={onFormReset}>
            <RepoSettingSection repositories={repositories} repoQuery={repoQuery} repoMode={repoMode} />

            <hr aria-hidden={true} className="my-4 w-100" />

            <FormGroup
                name="data series group"
                title="Data series"
                subtitle={
                    licensed ? 'Add any number of data series to your chart' : 'Add up to 10 data series to your chart'
                }
                error={(series.meta.touched && series.meta.error) || undefined}
                innerRef={series.input.ref}
            >
                <FormSeries
                    seriesField={series}
                    // Set repo query to preview only when search query mode is activated
                    repoQuery={repoMode.input.value === 'search-query' ? repoQuery.input.value.query : null}
                    repositories={repositories.input.value}
                    showValidationErrorsOnMount={submitted}
                />
            </FormGroup>

            <hr aria-hidden={true} className="my-4 w-100" />

            <FormGroup name="chart settings group" title="Chart settings">
                <Input
                    label="Title"
                    required={true}
                    message="Shown as the title for your insight"
                    placeholder="Example: Migration to React function components"
                    className="d-flex flex-column"
                    {...getDefaultInputProps(title)}
                />

                <CodeInsightTimeStepPicker
                    {...stepValue.input}
                    valid={stepValue.meta.touched && stepValue.meta.validState === 'VALID'}
                    error={(stepValue.meta.touched && stepValue.meta.error) || undefined}
                    errorInputState={stepValue.meta.touched && stepValue.meta.validState === 'INVALID'}
                    stepType={step.input.value}
                    onStepTypeChange={step.input.onChange}
                    numberOfPoints={12}
                />
            </FormGroup>

            {!!dashboardReferenceCount && dashboardReferenceCount > 1 && (
                <CodeInsightDashboardsVisibility className="mt-5 mb-n1" dashboardCount={dashboardReferenceCount} />
            )}

            <hr aria-hidden={true} className="my-4 w-100" />

            {children({ submitting, submitErrors, isFormClearActive })}
        </form>
    )
}
