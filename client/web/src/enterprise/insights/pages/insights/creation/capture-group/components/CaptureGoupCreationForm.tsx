import type { FC, FormHTMLAttributes, ReactNode } from 'react'

import classNames from 'classnames'

import {
    Card,
    Input,
    Label,
    Link,
    FormGroup,
    getDefaultInputProps,
    type useFieldAPI,
    type FormInstance,
    type SubmissionErrors,
} from '@sourcegraph/wildcard'

import {
    CodeInsightTimeStepPicker,
    CodeInsightDashboardsVisibility,
    LimitedAccessLabel,
    RepoSettingSection,
} from '../../../../../components'
import { useUiFeatures } from '../../../../../hooks'
import type { CaptureGroupFormFields } from '../types'
import type { Checks } from '../utils/search-query-validator'

import { CaptureGroupSeriesInfoBadge } from './info-badge/CaptureGroupSeriesInfoBadge'
import { CaptureGroupQueryInput } from './query-input/CaptureGroupQueryInput'
import { SearchQueryChecks } from './search-query-checks/SearchQueryChecks'

interface CaptureGroupCreationFormProps extends Omit<FormHTMLAttributes<HTMLFormElement>, 'title' | 'children'> {
    form: FormInstance<CaptureGroupFormFields>
    title: useFieldAPI<CaptureGroupFormFields['title']>
    repositories: useFieldAPI<CaptureGroupFormFields['repositories']>
    repoQuery: useFieldAPI<CaptureGroupFormFields['repoQuery']>
    repoMode: useFieldAPI<CaptureGroupFormFields['repoMode']>
    step: useFieldAPI<CaptureGroupFormFields['step']>
    stepValue: useFieldAPI<CaptureGroupFormFields['stepValue']>
    query: useFieldAPI<CaptureGroupFormFields['groupSearchQuery'], Checks>

    dashboardReferenceCount?: number
    isFormClearActive: boolean
    children: (inputs: RenderPropertyInputs) => ReactNode

    onFormReset: () => void
}

export interface RenderPropertyInputs {
    submitting: boolean
    submitErrors: SubmissionErrors
    isFormClearActive: boolean
}

export const CaptureGroupCreationForm: FC<CaptureGroupCreationFormProps> = props => {
    const {
        form,
        title,
        repoMode,
        repoQuery,
        repositories,
        query,
        step,
        stepValue,
        dashboardReferenceCount,
        isFormClearActive,
        children,
        onFormReset,
        ...attributes
    } = props

    const {
        handleSubmit,
        formAPI: { submitErrors, submitting },
    } = form
    const { licensed } = useUiFeatures()

    return (
        // eslint-disable-next-line react/forbid-elements
        <form {...attributes} noValidate={true} onSubmit={handleSubmit} onReset={onFormReset}>
            <RepoSettingSection repositories={repositories} repoQuery={repoQuery} repoMode={repoMode} />

            <hr aria-hidden={true} className="my-4 w-100" />

            <FormGroup
                name="data series"
                title="Data series"
                subtitle={
                    <>
                        Generated dynamically for each unique value from the regular expression capture group.{' '}
                        <Link
                            to="/help/code_insights/explanations/automatically_generated_data_series"
                            target="_blank"
                            rel="noopener"
                        >
                            Learn more.
                        </Link>
                    </>
                }
            >
                <Card className="p-3">
                    <Label className="w-100" id="capture-group-query-label">
                        <div className="mb-2">Search query</div>

                        <small className={classNames('mb-3', 'text-muted', 'd-block', 'font-weight-normal')}>
                            Search query must contain a properly formatted regular expression with at least one{' '}
                            <Link
                                to="/help/code_insights/explanations/automatically_generated_data_series#regular-expression-capture-group-resources"
                                target="_blank"
                                rel="noopener"
                            >
                                capture group.
                            </Link>{' '}
                            The capture group cannot match file or repository names, it can match only the file
                            contents.
                        </small>

                        <Input
                            as={CaptureGroupQueryInput}
                            required={true}
                            // Set repo query to preview only when search query mode is activated
                            repoQuery={repoMode.input.value === 'search-query' ? repoQuery.input.value.query : null}
                            repositories={repositories.input.value}
                            placeholder="Example: file:\.pom$ <java\.version>(.*)</java\.version>"
                            aria-labelledby="capture-group-query-label"
                            {...getDefaultInputProps(query)}
                        />
                    </Label>

                    <SearchQueryChecks checks={query.meta.validationContext} />

                    {!licensed && (
                        <LimitedAccessLabel message="Unlock Code Insights for unlimited data series" className="my-3" />
                    )}

                    <CaptureGroupSeriesInfoBadge>
                        <b className="font-weight-medium">Name</b> and <b className="font-weight-medium">color</b> of
                        each data series will be generated automatically. Chart will display{' '}
                        <b className="font-weight-medium">up to {licensed ? '20' : '10'}</b> data series.
                    </CaptureGroupSeriesInfoBadge>

                    <small className="mt-3">
                        Explore{' '}
                        <Link
                            to="/help/code_insights/references/common_use_cases#automatic-version-and-pattern-tracking"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            example queries
                        </Link>{' '}
                        and learn more about{' '}
                        <Link
                            to="/help/code_insights/explanations/automatically_generated_data_series"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            automatically generated data series
                        </Link>
                    </small>
                </Card>
            </FormGroup>

            <hr aria-hidden={true} className="my-4 w-100" />

            <FormGroup name="chart settings group" title="Chart settings">
                <Input
                    label="Title"
                    required={true}
                    message="Shown as the title for your insight"
                    placeholder="Example: Migration to React function components"
                    {...getDefaultInputProps(title)}
                    className="d-flex flex-column"
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
