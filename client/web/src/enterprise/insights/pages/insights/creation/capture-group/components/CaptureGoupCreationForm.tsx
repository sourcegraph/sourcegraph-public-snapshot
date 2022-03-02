import classNames from 'classnames'
import React from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Card, Link } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../components/LoaderButton'
import { CodeInsightTimeStepPicker, CodeInsightDashboardsVisibility } from '../../../../../components/creation-ui-kit'
import { FormGroup } from '../../../../../components/form/form-group/FormGroup'
import { FormInput } from '../../../../../components/form/form-input/FormInput'
import { useFieldAPI } from '../../../../../components/form/hooks/useField'
import { Form, FORM_ERROR } from '../../../../../components/form/hooks/useForm'
import { RepositoriesField } from '../../../../../components/form/repositories-field/RepositoriesField'
import { CaptureGroupFormFields } from '../types'
import { searchQueryValidator } from '../utils/search-query-validator'

import { CaptureGroupSeriesInfoBadge } from './info-badge/CaptureGroupSeriesInfoBadge'
import { CaptureGroupQueryInput } from './query-input/CaptureGroupQueryInput'
import { SearchQueryChecks } from './search-query-checks/SearchQueryChecks'

interface CaptureGroupCreationFormProps {
    mode: 'creation' | 'edit'
    form: Form<CaptureGroupFormFields>
    title: useFieldAPI<CaptureGroupFormFields['title']>
    repositories: useFieldAPI<CaptureGroupFormFields['repositories']>
    allReposMode: useFieldAPI<CaptureGroupFormFields['allRepos']>
    step: useFieldAPI<CaptureGroupFormFields['step']>
    stepValue: useFieldAPI<CaptureGroupFormFields['stepValue']>
    query: useFieldAPI<CaptureGroupFormFields['groupSearchQuery']>

    dashboardReferenceCount?: number
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
        allReposMode,
        query,
        step,
        stepValue,
        mode,
        dashboardReferenceCount,
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
        // eslint-disable-next-line react/forbid-elements
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

                <label className="d-flex flex-wrap align-items-center mb-2 mt-3 font-weight-normal">
                    <input
                        type="checkbox"
                        {...allReposMode.input}
                        value="all-repos-mode"
                        checked={allReposMode.input.value}
                    />

                    <span className="pl-2">Run your insight over all your repositories</span>

                    <small className="w-100 mt-2 text-muted">
                        This feature is actively in development. Read about the{' '}
                        <Link
                            to="/help/code_insights/explanations/current_limitations_of_code_insights"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            limitations here.
                        </Link>
                    </small>
                </label>
            </FormGroup>

            <hr className="my-4 w-100" />

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
                    <FormInput
                        title="Search query"
                        required={true}
                        as={CaptureGroupQueryInput}
                        repositories={repositories.input.value}
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
                    numberOfPoints={allReposMode.input.value ? 12 : 7}
                />
            </FormGroup>

            {!!dashboardReferenceCount && dashboardReferenceCount > 1 && (
                <CodeInsightDashboardsVisibility className="mt-5 mb-n1" dashboardCount={dashboardReferenceCount} />
            )}

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
                    className="mr-2 mb-2"
                    variant="primary"
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
        <Link
            to="/help/code_insights/explanations/automatically_generated_data_series#regular-expression-capture-group-resources"
            target="_blank"
            rel="noopener"
        >
            capture group.
        </Link>{' '}
        The capture group cannot match file or repository names, it can match only the file contents.
    </small>
)
