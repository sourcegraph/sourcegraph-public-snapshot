import classNames from 'classnames'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../../../../../components/alerts'
import { LoaderButton } from '../../../../../../../components/LoaderButton'
import { CodeInsightTimeStepPicker } from '../../../../../components/creation-ui-kit'
import { FormGroup } from '../../../../../components/form/form-group/FormGroup'
import { FormInput } from '../../../../../components/form/form-input/FormInput'
import { useFieldAPI } from '../../../../../components/form/hooks/useField'
import { Form, FORM_ERROR } from '../../../../../components/form/hooks/useForm'
import { RepositoriesField } from '../../../../../components/form/repositories-field/RepositoriesField'
import { LinkWithQuery } from '../../../../../components/link-with-query'
import { searchQueryValidator } from '../search-query-validator'
import { CaptureGroupFormFields } from '../types'

import { CaptureGroupSeriesInfoBadge } from './info-badge/CaptureGroupSeriesInfoBadge'
import { CaptureGroupQueryInput } from './query-input/CaptureGroupQueryInput'
import { SearchQueryChecks } from './search-query-checks/SearchQueryChecks'

interface CaptureGroupCreationFormProps {
    mode: 'creation' | 'edit'
    form: Form<CaptureGroupFormFields>
    title: useFieldAPI<CaptureGroupFormFields['title']>
    repositories: useFieldAPI<CaptureGroupFormFields['repositories']>
    step: useFieldAPI<CaptureGroupFormFields['step']>
    stepValue: useFieldAPI<CaptureGroupFormFields['stepValue']>
    query: useFieldAPI<CaptureGroupFormFields['groupSearchQuery']>

    isFormClearActive?: boolean
    className?: string

    onCancel: () => void
    onFormReset: () => void
}

export const CaptureGroupCreationForm: React.FunctionComponent<CaptureGroupCreationFormProps> = props => {
    const {
        form,
        title,
        repositories,
        query,
        step,
        stepValue,
        mode,
        className,
        isFormClearActive,
        onFormReset,
        onCancel,
    } = props

    const {
        ref,
        handleSubmit,
        formAPI: { submitErrors, submitting },
    } = form
    const isEditMode = mode === 'edit'

    return (
        <form noValidate={true} ref={ref} className={className} onSubmit={handleSubmit} onReset={onFormReset}>
            <FormGroup
                name="insight repositories"
                title="Targeted repositories"
                subtitle="Create a list of repositories to run your search over"
            >
                <FormInput
                    as={RepositoriesField}
                    autoFocus={true}
                    required={true}
                    title="Repositories"
                    description="Separate repositories with commas"
                    placeholder="Example: github.com/sourcegraph/sourcegraph"
                    loading={repositories.meta.validState === 'CHECKING'}
                    valid={repositories.meta.touched && repositories.meta.validState === 'VALID'}
                    error={repositories.meta.touched && repositories.meta.error}
                    {...repositories.input}
                    className="mb-0 d-flex flex-column"
                />

                <div className="d-flex mb-2 mt-3 align-items-start">
                    <InformationOutlineIcon className="text-muted pr-2 h-auto flex-shrink-0" />

                    <small className="text-muted">
                        This type of insight can only run across specified repositories. To run your insight across all
                        repositories, use <LinkWithQuery to="/insights/create/search">"Track" insight</LinkWithQuery>{' '}
                        and define data series manually. Learn about the{' '}
                        <a
                            href="https://docs.sourcegraph.com/code_insights/explanations/current_limitations_of_code_insights"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            beta limitations.
                        </a>
                    </small>
                </div>
            </FormGroup>

            <hr className="my-4 w-100" />

            <FormGroup
                name="data series"
                title="Data series"
                subtitle={
                    <>
                        Generated dynamically for each unique value from the regular expression capture group.{' '}
                        <a
                            href="https://docs.sourcegraph.com/code_insights/explanations/automatically_generated_data_series"
                            target="_blank"
                            rel="noopener"
                        >
                            Learn more.
                        </a>
                    </>
                }
            >
                <div className="card card-body p-3">
                    <FormInput
                        title="Search query"
                        required={true}
                        as={CaptureGroupQueryInput}
                        subtitle={<QueryFieldSubtitle className="mb-3" />}
                        placeholder="Example: file:\.pom$ <java\.version>(.*)</java\.version>"
                        valid={query.meta.touched && query.meta.validState === 'VALID'}
                        error={query.meta.touched && query.meta.error}
                        className="mb-3"
                        {...query.input}
                    />

                    <SearchQueryChecks checks={searchQueryValidator(query.input.value, query.meta.touched)} />

                    <CaptureGroupSeriesInfoBadge>
                        <b className="font-weight-medium">Name</b> and <b className="font-weight-medium">color</b> of
                        each data series will be generated automatically. Chart will display{' '}
                        <b className="font-weight-medium">up to 20</b> data series.
                    </CaptureGroupSeriesInfoBadge>

                    <small className="mt-3">
                        Explore{' '}
                        <a
                            href="https://docs.sourcegraph.com/code_insights/references/common_use_cases#automatic-version-and-pattern-tracking"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            example queries
                        </a>{' '}
                        and learn more about{' '}
                        <a
                            href="https://docs.sourcegraph.com/code_insights/explanations/automatically_generated_data_series"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            automatically generated data series
                        </a>
                    </small>
                </div>
            </FormGroup>

            <hr className="my-4 w-100" />

            <FormGroup name="chart settings group" title="Chart settings">
                <FormInput
                    title="Title"
                    required={true}
                    description="Shown as the title for your insight"
                    placeholder="Example: Migration to React function components"
                    valid={title.meta.touched && title.meta.validState === 'VALID'}
                    error={title.meta.touched && title.meta.error}
                    {...title.input}
                    className="d-flex flex-column"
                />

                <CodeInsightTimeStepPicker
                    {...stepValue.input}
                    valid={stepValue.meta.touched && stepValue.meta.validState === 'VALID'}
                    error={stepValue.meta.touched && stepValue.meta.error}
                    errorInputState={stepValue.meta.touched && stepValue.meta.validState === 'INVALID'}
                    stepType={step.input.value}
                    onStepTypeChange={step.input.onChange}
                />
            </FormGroup>

            <hr className="my-4 w-100" />

            <footer className="d-flex flex-wrap align-items-center">
                {submitErrors?.[FORM_ERROR] && <ErrorAlert className="w-100" error={submitErrors[FORM_ERROR]} />}

                <LoaderButton
                    type="submit"
                    alwaysShowLabel={true}
                    loading={submitting}
                    label={submitting ? 'Submitting' : isEditMode ? 'Save insight' : 'Create code insight'}
                    disabled={submitting}
                    data-testid="insight-save-button"
                    className="btn btn-primary mr-2 mb-2"
                />

                <Button type="button" variant="secondary" outline={true} className="mb-2 mr-auto" onClick={onCancel}>
                    Cancel
                </Button>

                <Button
                    type="reset"
                    disabled={!isFormClearActive}
                    variant="secondary"
                    outline={true}
                    className="border-0"
                >
                    Clear all fields
                </Button>
            </footer>
        </form>
    )
}

const QueryFieldSubtitle: React.FunctionComponent<{ className?: string }> = props => (
    <small className={classNames(props.className, 'text-muted', 'd-block', 'font-weight-normal')}>
        Search query must contain a properly formatted regular expression with at least one{' '}
        <a
            href="https://docs.sourcegraph.com/code_insights/explanations/automatically_generated_data_series#regular-expression-capture-group-resources"
            target="_blank"
            rel="noopener"
        >
            capture group.
        </a>{' '}
        The capture group cannot match file or repository names, it can match only the file contents.
    </small>
)
