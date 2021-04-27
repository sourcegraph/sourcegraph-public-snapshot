import classnames from 'classnames'
import { FORM_ERROR, FormApi, SubmissionErrors } from 'final-form'
import createFocusDecorator from 'final-form-focus'
import React, { useEffect, useMemo, useRef } from 'react'
import { useField, useForm } from 'react-final-form-hooks'
import { noop } from 'rxjs'

import { Page } from '../../../../../components/Page'
import { PageTitle } from '../../../../../components/PageTitle'
import { DataSeries } from '../../types'
import { InputField } from '../form-field/FormField'
import { FormGroup } from '../form-group/FormGroup'
import { FormRadioInput } from '../form-radio-input/FormRadioInput'
import { FormSeries, FormSeriesReferenceAPI } from '../form-series/FormSeries'
import { createRequiredValidator, composeValidators, ValidationResult } from '../validators'

import styles from './CreateInsightForm.module.scss'

const requiredTitleField = createRequiredValidator('Title is required field for code insight.')
const repositoriesFieldValidator = composeValidators(
    createRequiredValidator('Repositories is required field for code insight.')
)

const requiredStepValueField = createRequiredValidator('Please specify a step between points.')
const seriesRequired = (series: DataSeries[]): ValidationResult =>
    series && series.length > 0 ? undefined : 'Series is empty. You must have at least one series for code insight.'

const INITIAL_VALUES: Partial<CreateInsightFormFields> = {
    visibility: 'personal',
    series: [],
    step: 'months',
}

/** Public API of code insight creation form. */
export interface CreateInsightFormProps {
    onSubmit: (
        values: CreateInsightFormFields,
        form: FormApi<CreateInsightFormFields, Partial<CreateInsightFormFields>>
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
    stepValue: number
}

/** Displays creation code insight form (title, visibility, series, etc.) */
export const CreateInsightForm: React.FunctionComponent<CreateInsightFormProps> = props => {
    const { onSubmit } = props

    const titleReference = useRef<HTMLInputElement>(null)
    const repositoriesReference = useRef<HTMLInputElement>(null)
    const seriesReference = useRef<FormSeriesReferenceAPI>(null)
    const stepValueReference = useRef<HTMLInputElement>(null)

    const focusOnErrorsDecorator = useMemo(() => {
        const noopFocus = { focus: noop, name: '' }

        return createFocusDecorator<CreateInsightFormFields>(() => [
            titleReference.current ?? noopFocus,
            repositoriesReference.current ?? noopFocus,
            seriesReference.current ?? noopFocus,
            stepValueReference.current ?? noopFocus,
        ])
    }, [])

    const { form, handleSubmit, submitErrors } = useForm<CreateInsightFormFields>({
        initialValues: INITIAL_VALUES,
        onSubmit,
    })

    useEffect(() => focusOnErrorsDecorator(form), [form, focusOnErrorsDecorator])

    const title = useField('title', form, requiredTitleField)
    const repositories = useField('repositories', form, repositoriesFieldValidator)
    const visibility = useField('visibility', form)
    const series = useField<DataSeries[], CreateInsightFormFields>('series', form, seriesRequired)
    const step = useField('step', form)
    const stepValue = useField('stepValue', form, requiredStepValueField)

    return (
        <Page className="col-8">
            <PageTitle title="Create new code insight" />

            <div className={styles.createInsightFormSubTitleContainer}>
                <h2>Create new code insight</h2>

                <p className="text-muted">
                    Search-based code insights analyse your code based on any search query.{' '}
                    <a
                        href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points"
                        target="_blank"
                        rel="noopener"
                    >
                        Learn more.
                    </a>
                </p>
            </div>

            {/* eslint-disable-next-line react/forbid-elements */}
            <form onSubmit={handleSubmit} className={styles.createInsightFormForm}>
                <InputField
                    title="Title"
                    autofocus={true}
                    description="Shown as title for your insight"
                    placeholder="ex. Migration to React function components"
                    valid={title.meta.touched && title.meta.valid}
                    error={title.meta.touched && title.meta.error}
                    {...title.input}
                    ref={titleReference}
                    className={styles.createInsightFormFormField}
                />

                <InputField
                    title="Repositories"
                    description="Create a list of repositories to run your search over. Separate them with comas."
                    placeholder="Add or search for repositories"
                    valid={repositories.meta.touched && repositories.meta.valid}
                    error={repositories.meta.touched && repositories.meta.error}
                    {...repositories.input}
                    ref={repositoriesReference}
                    className={styles.createInsightFormFormField}
                />

                <FormGroup
                    name="visibility"
                    title="Visibility"
                    description="This insigh will be visible only on your personal dashboard. It will not be show to other
                        users in your organisation."
                    className={styles.createInsightFormFormField}
                >
                    <div className={styles.createInsightFormRadioGroupContent}>
                        <FormRadioInput
                            name="visibility"
                            value="personal"
                            title="Personal"
                            description="only for you"
                            checked={visibility.input.value === 'personal'}
                            className={styles.createInsightFormRadio}
                            onChange={visibility.input.onChange}
                        />

                        <FormRadioInput
                            name="visibility"
                            value="organization"
                            title="Organization"
                            description="to all users in your organization"
                            checked={visibility.input.value === 'organization'}
                            onChange={visibility.input.onChange}
                            className={styles.createInsightFormRadio}
                        />
                    </div>
                </FormGroup>

                <FormGroup
                    name="data series group"
                    title="Data series"
                    subtitle="Add any number of data series to your chart"
                    error={series.meta.touched && series.meta.error}
                    className={styles.createInsightFormFormField}
                >
                    <FormSeries
                        name={series.input.name}
                        ref={seriesReference}
                        series={series.input.value}
                        onChange={series.input.onChange}
                    />
                </FormGroup>

                <FormGroup
                    name="insight step group"
                    title="Step between data points"
                    description="The distance between two data points on the chart"
                    error={stepValue.meta.touched && stepValue.meta.error}
                    className={styles.createInsightFormFormField}
                >
                    <div className={styles.createInsightFormRadioGroupContent}>
                        <InputField
                            placeholder="ex. 2"
                            {...stepValue.input}
                            valid={stepValue.meta.touched && stepValue.meta.valid}
                            ref={stepValueReference}
                            className={classnames(styles.createInsightFormStepInput)}
                        />

                        <FormRadioInput
                            title="Hours"
                            name="step"
                            value="hours"
                            checked={step.input.value === 'hours'}
                            onChange={step.input.onChange}
                            className={styles.createInsightFormRadio}
                        />
                        <FormRadioInput
                            title="Days"
                            name="step"
                            value="days"
                            checked={step.input.value === 'days'}
                            onChange={step.input.onChange}
                            className={styles.createInsightFormRadio}
                        />
                        <FormRadioInput
                            title="Weeks"
                            name="step"
                            value="weeks"
                            checked={step.input.value === 'weeks'}
                            onChange={step.input.onChange}
                            className={styles.createInsightFormRadio}
                        />
                        <FormRadioInput
                            title="Months"
                            name="step"
                            value="months"
                            checked={step.input.value === 'months'}
                            onChange={step.input.onChange}
                            className={styles.createInsightFormRadio}
                        />
                        <FormRadioInput
                            title="Years"
                            name="step"
                            value="years"
                            checked={step.input.value === 'years'}
                            onChange={step.input.onChange}
                            className={styles.createInsightFormRadio}
                        />
                    </div>
                </FormGroup>

                <div className={styles.createInsightFormButtons}>
                    {submitErrors?.[FORM_ERROR] && (
                        <div className="alert alert-danger">{submitErrors[FORM_ERROR].toString()}</div>
                    )}

                    <button
                        type="submit"
                        className={classnames(styles.createInsightFormButton, styles.createInsightFormButtonActive)}
                    >
                        Create code insight
                    </button>
                    <button type="button" className={classnames(styles.createInsightFormButton)}>
                        Cancel
                    </button>
                </div>
            </form>
        </Page>
    )
}
