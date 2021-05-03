import classnames from 'classnames'
import React from 'react'

import { FORM_ERROR, SubmissionErrors, useField, useForm, Validator } from '../../hooks/useForm';
import { DataSeries } from '../../types'
import { InputField } from '../form-field/FormField'
import { FormGroup } from '../form-group/FormGroup'
import { FormRadioInput } from '../form-radio-input/FormRadioInput'
import { FormSeries } from '../form-series/FormSeries'
import { createRequiredValidator } from '../validators'

import styles from './CreationSearchInsightForm.module.scss'

const requiredTitleField = createRequiredValidator('Title is required field for code insight.')
const repositoriesFieldValidator = createRequiredValidator('Repositories is required field for code insight.')

const requiredStepValueField = createRequiredValidator('Please specify a step between points.')
const seriesRequired: Validator<DataSeries[]> = series =>
    series && series.length > 0
        ? undefined
        : 'Series is empty. You must have at least one series for code insight.'

const INITIAL_VALUES: Partial<CreateInsightFormFields> = {
    visibility: 'personal',
    series: [],
    step: 'months',
}

/** Public API of code insight creation form. */
export interface CreationSearchInsightFormProps {
    /** Custom class name for root form element. */
    className?: string
    /** Submit handler for form element. */
    onSubmit: (
        values: CreateInsightFormFields,
    ) => SubmissionErrors | Promise<SubmissionErrors> | void
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
export const CreationSearchInsightForm: React.FunctionComponent<CreationSearchInsightFormProps> = props => {
    const { className, onSubmit } = props

    const { formAPI, ref, handleSubmit } = useForm<CreateInsightFormFields>({
        initialValues: INITIAL_VALUES,
        onSubmit
    })

    const title = useField('title', formAPI, requiredTitleField)
    const repositories = useField('repositories', formAPI, repositoriesFieldValidator)
    const visibility = useField('visibility', formAPI)

    const series = useField('series', formAPI, seriesRequired)
    const step = useField('step', formAPI)
    const stepValue = useField('stepValue', formAPI, requiredStepValueField)

    return (
        // eslint-disable-next-line react/forbid-elements
        <form noValidate={true} ref={ref} onSubmit={handleSubmit} className={classnames(className, 'd-flex flex-column')}>
            <InputField
                title="Title"
                required={true}
                description="Shown as title for your insight"
                placeholder="ex. Migration to React function components"
                valid={title.meta.touched && title.meta.validState === 'VALID'}
                error={title.meta.touched && title.meta.error}
                {...title.input}
                className="mb-0"
            />

            <InputField
                title="Repositories"
                required={true}
                description="Create a list of repositories to run your search over. Separate them with comas."
                placeholder="Add or search for repositories"
                valid={repositories.meta.touched && repositories.meta.validState === 'VALID'}
                error={repositories.meta.touched && repositories.meta.error}
                {...repositories.input}
                className="mb-0 mt-4"
            />

            <FormGroup
                name="visibility"
                title="Visibility"
                description="This insigh will be visible only on your personal dashboard. It will not be show to other
                            users in your organisation."
                className="mb-0 mt-4"
                contentClassName="d-flex flex-wrap mb-n2"
            >
                <FormRadioInput
                    name="visibility"
                    value="personal"
                    title="Personal"
                    description="only for you"
                    checked={visibility.input.value === 'personal'}
                    className="mr-3"
                    onChange={visibility.input.onChange}
                />

                <FormRadioInput
                    name="visibility"
                    value="organization"
                    title="Organization"
                    description="to all users in your organization"
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
                <FormSeries
                    series={series.input.value}
                    onChange={series.input.onChange}
                />
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
                <InputField
                    placeholder="ex. 2"
                    required={true}
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

                {formAPI.submitErrors?.[FORM_ERROR] && (
                    <div className="alert alert-danger">{formAPI.submitErrors[FORM_ERROR]}</div>
                )}

                <button type="submit" className="btn btn-primary mr-2">
                    Create code insight
                </button>
                <button type="button" className="btn btn-outline-secondary">
                    Cancel
                </button>
            </div>
        </form>
    )
}
