import { FC, FormEventHandler, RefObject, useMemo } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Checkbox, Input, Link, useObservable } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../components/LoaderButton'
import {
    FormSeries,
    LimitedAccessLabel,
    CodeInsightDashboardsVisibility,
    CodeInsightTimeStepPicker,
    RepositoriesField,
    FormGroup,
    getDefaultInputProps,
    useFieldAPI,
    FORM_ERROR,
    SubmissionErrors,
} from '../../../../../components'
import { Insight } from '../../../../../core'
import { useUiFeatures } from '../../../../../hooks'
import { CreateInsightFormFields } from '../types'

interface CreationSearchInsightFormProps {
    /** This component might be used in edit or creation insight case. */
    mode?: 'creation' | 'edit'

    innerRef: RefObject<any>
    handleSubmit: FormEventHandler
    submitErrors: SubmissionErrors
    submitting: boolean
    submitted: boolean
    className?: string
    isFormClearActive?: boolean
    dashboardReferenceCount?: number

    title: useFieldAPI<CreateInsightFormFields['title']>
    repositories: useFieldAPI<CreateInsightFormFields['repositories']>
    allReposMode: useFieldAPI<CreateInsightFormFields['allRepos']>

    series: useFieldAPI<CreateInsightFormFields['series']>
    step: useFieldAPI<CreateInsightFormFields['step']>
    stepValue: useFieldAPI<CreateInsightFormFields['stepValue']>
    insight?: Insight

    onCancel: () => void
    onFormReset: () => void
}

/**
 * Displays creation code insight form (title, visibility, series, etc.)
 * UI layer only, all controlled data should be managed by consumer of this component.
 */
export const SearchInsightCreationForm: FC<CreationSearchInsightFormProps> = props => {
    const {
        mode,
        innerRef,
        handleSubmit,
        submitErrors,
        submitting,
        submitted,
        title,
        repositories,
        allReposMode,
        series,
        stepValue,
        step,
        className,
        isFormClearActive,
        dashboardReferenceCount,
        insight,
        onCancel,
        onFormReset,
    } = props

    const isEditMode = mode === 'edit'
    const { licensed, insight: insightFeatures } = useUiFeatures()

    const creationPermission = useObservable(
        useMemo(
            () =>
                isEditMode && insight
                    ? insightFeatures.getEditPermissions(insight)
                    : insightFeatures.getCreationPermissions(),
            [insightFeatures, isEditMode, insight]
        )
    )

    return (
        // eslint-disable-next-line react/forbid-elements
        <form noValidate={true} ref={innerRef} onSubmit={handleSubmit} onReset={onFormReset} className={className}>
            <FormGroup
                name="insight repositories"
                title="Targeted repositories"
                subtitle="Create a list of repositories to run your search over"
            >
                <Input
                    as={RepositoriesField}
                    autoFocus={true}
                    required={true}
                    label="Repositories"
                    message="Separate repositories with commas"
                    placeholder={
                        allReposMode.input.value ? 'All repositories' : 'Example: github.com/sourcegraph/sourcegraph'
                    }
                    className="mb-0 d-flex flex-column"
                    {...getDefaultInputProps(repositories)}
                />

                <Checkbox
                    {...allReposMode.input}
                    type="checkbox"
                    id="RunInsightsOnAllRepoCheck"
                    wrapperClassName="mb-1 mt-3 font-weight-normal"
                    value="all-repos-mode"
                    checked={allReposMode.input.value}
                    label="Run your insight over all your repositories"
                />

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

                <hr className="my-4 w-100" />
            </FormGroup>

            <FormGroup
                name="data series group"
                title="Data series"
                subtitle={
                    licensed ? 'Add any number of data series to your chart' : 'Add up to 10 data series to your chart'
                }
                error={series.meta.touched && series.meta.error}
                innerRef={series.input.ref}
            >
                <FormSeries
                    seriesField={series}
                    repositories={repositories.input.value}
                    showValidationErrorsOnMount={submitted}
                />
            </FormGroup>

            <hr className="my-4 w-100" />

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

            {!licensed && (
                <LimitedAccessLabel message="Unlock Code Insights to create unlimited insights" className="mb-3" />
            )}

            <div className="d-flex flex-wrap align-items-center">
                {submitErrors?.[FORM_ERROR] && <ErrorAlert className="w-100" error={submitErrors[FORM_ERROR]} />}

                <LoaderButton
                    alwaysShowLabel={true}
                    loading={submitting}
                    label={submitting ? 'Submitting' : isEditMode ? 'Save changes' : 'Create code insight'}
                    type="submit"
                    disabled={submitting || !creationPermission?.available}
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
            </div>
        </form>
    )
}
