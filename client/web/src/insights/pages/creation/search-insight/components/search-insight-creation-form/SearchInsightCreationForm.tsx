import classnames from 'classnames'
import React from 'react'
import { noop } from 'rxjs'

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { ErrorAlert } from '../../../../../../components/alerts'
import { LoaderButton } from '../../../../../../components/LoaderButton'
import { FormGroup } from '../../../../../components/form/form-group/FormGroup'
import { FormInput } from '../../../../../components/form/form-input/FormInput'
import { FormRadioInput } from '../../../../../components/form/form-radio-input/FormRadioInput'
import { useField, Validator } from '../../../../../components/form/hooks/useField'
import { FORM_ERROR, SubmissionErrors, useForm } from '../../../../../components/form/hooks/useForm'
import { useTitleValidator } from '../../../../../components/form/hooks/useTitleValidator'
import { createRequiredValidator } from '../../../../../components/form/validators'
import { InsightTypeSuffix } from '../../../../../core/types'
import { DataSeries } from '../../types'
import { FormSeries } from '../form-series/FormSeries'

import styles from './SearchInsightCreationForm.module.scss'

const repositoriesFieldValidator = createRequiredValidator('Repositories is a required field.')
const requiredStepValueField = createRequiredValidator('Please specify a step between points.')
/**
 * Custom validator for chart series. Since series has complex type
 * we can't validate this with standard validators.
 * */
const seriesRequired: Validator<DataSeries[]> = series =>
    series && series.length > 0 ? undefined : 'Series is empty. You must have at least one series for code insight.'

const INITIAL_VALUES: Partial<CreateInsightFormFields> = {
    visibility: 'personal',
    series: [],
    step: 'months',
    stepValue: '2',
    title: '',
    repositories: '',
}

/** Public API of code insight creation form. */
export interface CreationSearchInsightFormProps {
    /** Final settings cascade. Used for title field validation. */
    settings?: Settings | null
    /** Custom class name for root form element. */
    className?: string
    /** Submit handler for form element. */
    onSubmit: (values: CreateInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel?: () => void
}

/** Creation form fields. */
export interface CreateInsightFormFields {
    /** Code Insight series setting (name of line, line query, color) */
    series: DataSeries[]
    /** Title of code insight*/
    title: string
    /** Repositories which to be used to get the info for code insights */
    repositories: string
    /** Visibility setting which responsible for where insight will appear. */
    visibility: 'personal' | 'organization'
    /** Setting for set chart step - how often do we collect data. */
    step: 'hours' | 'days' | 'weeks' | 'months' | 'years'
    /** Value for insight step setting */
    stepValue: string
}

/** Displays creation code insight form (title, visibility, series, etc.) */
export const SearchInsightCreationForm: React.FunctionComponent<CreationSearchInsightFormProps> = props => {
    const { settings, className, onSubmit, onCancel = noop } = props

    const { formAPI, ref, handleSubmit } = useForm<CreateInsightFormFields>({
        initialValues: INITIAL_VALUES,
        onSubmit,
    })

    // We can't have two or more insights with the same name, since we rely on name as on id of insights.
    const titleValidator = useTitleValidator({ settings, insightType: InsightTypeSuffix.search })

    const title = useField('title', formAPI, titleValidator)
    const repositories = useField('repositories', formAPI, repositoriesFieldValidator)
    const visibility = useField('visibility', formAPI)

    const series = useField('series', formAPI, seriesRequired)
    const step = useField('step', formAPI)
    const stepValue = useField('stepValue', formAPI, requiredStepValueField)

    return (
        // eslint-disable-next-line react/forbid-elements
        <form
            noValidate={true}
            ref={ref}
            onSubmit={handleSubmit}
            className={classnames(className, 'd-flex flex-column')}
        >
            <FormInput
                title="Title"
                autoFocus={true}
                required={true}
                description="Shown as the title for your insight"
                placeholder="ex. Migration to React function components"
                valid={title.meta.touched && title.meta.validState === 'VALID'}
                error={title.meta.touched && title.meta.error}
                {...title.input}
                className="mb-0"
            />

            <FormInput
                title="Repositories"
                required={true}
                description="Create a list of repositories to run your search over. Separate them with commas."
                placeholder="Add or search for repositories"
                valid={repositories.meta.touched && repositories.meta.validState === 'VALID'}
                error={repositories.meta.touched && repositories.meta.error}
                {...repositories.input}
                className="mb-0 mt-4"
            />

            <FormGroup
                name="visibility"
                title="Visibility"
                description="This insight will be visible only on your personal dashboard. It will not be show to other
                            users in your organization."
                className="mb-0 mt-4"
                contentClassName="d-flex flex-wrap mb-n2"
            >
                <FormRadioInput
                    name="visibility"
                    value="personal"
                    title="Personal"
                    description="only you"
                    checked={visibility.input.value === 'personal'}
                    className="mr-3"
                    onChange={visibility.input.onChange}
                />

                <FormRadioInput
                    name="visibility"
                    value="organization"
                    title="Organization"
                    description="all users in your organization"
                    checked={visibility.input.value === 'organization'}
                    onChange={visibility.input.onChange}
                    className="mr-3"
                />
            </FormGroup>

            <hr className={styles.creationInsightFormSeparator} />

            <FormGroup
                name="data series group"
                title="Data series"
                subtitle="Add any number of data series to your chart"
                error={series.meta.touched && series.meta.error}
                innerRef={series.input.ref}
                className="mb-0"
            >
                <FormSeries series={series.input.value} onChange={series.input.onChange} />
            </FormGroup>

            <hr className={styles.creationInsightFormSeparator} />

            <FormGroup
                name="insight step group"
                title="Step between data points"
                description="The distance between two data points on the chart"
                error={stepValue.meta.touched && stepValue.meta.error}
                className="mb-0"
                contentClassName="d-flex flex-wrap mb-n2"
            >
                <FormInput
                    placeholder="ex. 2"
                    required={true}
                    type="number"
                    min={1}
                    {...stepValue.input}
                    valid={stepValue.meta.touched && stepValue.meta.validState === 'VALID'}
                    errorInputState={stepValue.meta.touched && stepValue.meta.validState === 'INVALID'}
                    className={classnames(styles.creationInsightFormStepInput)}
                />

                <FormRadioInput
                    title="Hours"
                    name="step"
                    value="hours"
                    checked={step.input.value === 'hours'}
                    onChange={step.input.onChange}
                    className="mr-3"
                />
                <FormRadioInput
                    title="Days"
                    name="step"
                    value="days"
                    checked={step.input.value === 'days'}
                    onChange={step.input.onChange}
                    className="mr-3"
                />
                <FormRadioInput
                    title="Weeks"
                    name="step"
                    value="weeks"
                    checked={step.input.value === 'weeks'}
                    onChange={step.input.onChange}
                    className="mr-3"
                />
                <FormRadioInput
                    title="Months"
                    name="step"
                    value="months"
                    checked={step.input.value === 'months'}
                    onChange={step.input.onChange}
                    className="mr-3"
                />
                <FormRadioInput
                    title="Years"
                    name="step"
                    value="years"
                    checked={step.input.value === 'years'}
                    onChange={step.input.onChange}
                    className="mr-3"
                />
            </FormGroup>

            <hr className={styles.creationInsightFormSeparator} />

            <div>
                {formAPI.submitErrors?.[FORM_ERROR] && <ErrorAlert error={formAPI.submitErrors[FORM_ERROR]} />}

                <LoaderButton
                    alwaysShowLabel={true}
                    loading={formAPI.submitting}
                    label={formAPI.submitting ? 'Submitting' : 'Create code insight'}
                    type="submit"
                    disabled={formAPI.submitting}
                    className="btn btn-primary mr-2"
                />

                <button type="button" className="btn btn-outline-secondary" onClick={onCancel}>
                    Cancel
                </button>
            </div>
        </form>
    )
}
