import classnames from 'classnames';
import createFocusDecorator from 'final-form-focus'
import React, { useEffect, useMemo, useRef } from 'react';
import { useField, useForm } from 'react-final-form-hooks';
import { noop } from 'rxjs';

import { Page } from '../../../components/Page';
import { PageTitle } from '../../../components/PageTitle';

import { InputField } from './components/form-field/FormField';
import { FormGroup } from './components/form-group/FormGroup';
import { FormRadioInput } from './components/form-radio-input/FormRadioInput';
import { FormSeries, FormSeriesReferenceAPI } from './components/form-series/FormSeries';
import { createRequiredValidator, composeValidators, ValidationResult } from './components/validators';
import styles from './CreateinsightPage.module.scss'
import { DataSeries } from './types';

interface CreateInsightPageProps {}

interface Form {
    series: DataSeries[];
    title: string;
    repositories: string;
    visibility: string;
}

const requiredTitleField = createRequiredValidator('Title is required field for code insight.');
const repositoriesFieldValidator = composeValidators(
    createRequiredValidator('Repositories is required field for code insight.'),
);

const requiredStepValueField = createRequiredValidator('Please specify a step between points.')
const seriesRequired = (series: DataSeries[]): ValidationResult =>
    series && series.length > 0 ? undefined : 'Series is empty. You must have at least one series for code insight.';

const INITIAL_VALUES = {
    visibility: 'personal',
    series: [],
    step: 'months',
}

export const CreateInsightPage: React.FunctionComponent<CreateInsightPageProps> = props => {
    const {} = props;

    const titleReference = useRef<HTMLInputElement>(null);
    const repositoriesReference = useRef<HTMLInputElement>(null);
    const seriesReference = useRef<FormSeriesReferenceAPI>(null);
    const stepValueReference = useRef<HTMLInputElement>(null);

    const focusOnErrorsDecorator = useMemo(() => {
        const noopFocus = { focus: noop, name: '' }

        return createFocusDecorator<Form>(
            () => [
                titleReference.current ?? noopFocus,
                repositoriesReference.current ?? noopFocus,
                seriesReference.current ?? noopFocus,
                stepValueReference.current ?? noopFocus,
            ],
        )
    }, []);

    const { form, handleSubmit } = useForm<Form>({
        initialValues: INITIAL_VALUES,
        onSubmit: (values, form) => {
            const { errors } = form.getState();

            console.log({errors});
        }
    });

    useEffect(() => focusOnErrorsDecorator(form), [form, focusOnErrorsDecorator])

    const title = useField('title', form, requiredTitleField);
    const repositories = useField('repositories', form, repositoriesFieldValidator);
    const visibility = useField('visibility', form);
    const series = useField<DataSeries[], Form>('series', form, seriesRequired);
    const step = useField('step', form);
    const stepValue = useField('stepValue', form, requiredStepValueField);

    return (
        <Page className='col-8'>
            <PageTitle title='Create new code insight'/>

            <div className={styles.createInsightSubTitleContainer}>

                <h2>Create new code insight</h2>

                <p className='text-muted'>
                    Search-based code insights analyse your code based on any search query.
                    {' '}
                    <a href="https://docs.sourcegraph.com/code_monitoring/how-tos/starting_points"
                       target="_blank"
                       rel="noopener">
                        Learn more.
                    </a>
                </p>
            </div>

            {/* eslint-disable-next-line react/forbid-elements */}
            <form onSubmit={handleSubmit} className={styles.createInsightForm}>

                <InputField
                    title='Title'
                    description='Shown as title for your insight'
                    placeholder='ex. Migration to React function components'
                    error={title.meta.touched && title.meta.error}
                    {...title.input}
                    ref={titleReference}
                    className={styles.createInsightFormField}/>

                <InputField
                    title='Repositories'
                    description='Create a list of repositories to run your search over. Separate them with comas.'
                    placeholder='Add or search for repositories'
                    error={repositories.meta.touched && repositories.meta.error}
                    {...repositories.input}
                    ref={repositoriesReference}
                    className={styles.createInsightFormField}/>

                <FormGroup
                    name='Visibility'
                    description='This insigh will be visible only on your personal dashboard. It will not be show to other
                        users in your organisation.'
                    className={styles.createInsightFormField}>

                    <div className={styles.createInsightRadioGroupContent}>

                        <FormRadioInput
                            name='visibility'
                            value='personal'
                            title='Personal'
                            description='only for you'
                            checked={visibility.input.value === 'personal'}
                            className={styles.createInsightRadio}
                            onChange={visibility.input.onChange}/>

                        <FormRadioInput
                            name='visibility'
                            value='organization'
                            title='Organization'
                            description='to all users in your organization'
                            checked={visibility.input.value === 'organization'}
                            onChange={visibility.input.onChange}
                            className={styles.createInsightRadio}/>
                    </div>
                </FormGroup>

                <FormGroup
                    name='Data series'
                    subtitle='Add any number of data series to your chart'
                    error={series.meta.touched && series.meta.error}
                    className={styles.createInsightFormField}>

                    <FormSeries
                        name={series.input.name}
                        ref={seriesReference}
                        series={series.input.value}
                        onChange={series.input.onChange}/>

                </FormGroup>

                <FormGroup
                    name='Step between data points'
                    description='The distance between two data points on the chart'
                    error={stepValue.meta.touched && stepValue.meta.error}
                    className={styles.createInsightFormField}>

                    <div className={styles.createInsightRadioGroupContent}>

                        <InputField
                            placeholder='ex. 2'
                            {...stepValue.input}
                            ref={stepValueReference}
                            className={classnames(styles.createInsightStepInput)}/>

                        <FormRadioInput
                            title='Hours'
                            name='step'
                            value='hours'
                            checked={step.input.value === 'hours'}
                            onChange={step.input.onChange}
                            className={styles.createInsightRadio}/>
                        <FormRadioInput
                            title='Days'
                            name='step'
                            value='days'
                            checked={step.input.value === 'days'}
                            onChange={step.input.onChange}
                            className={styles.createInsightRadio}/>
                        <FormRadioInput
                            title='Weeks'
                            name='step'
                            value='weeks'
                            checked={step.input.value === 'weeks'}
                            onChange={step.input.onChange}
                            className={styles.createInsightRadio}/>
                        <FormRadioInput
                            title='Months'
                            name='step'
                            value='months'
                            checked={step.input.value === 'months'}
                            onChange={step.input.onChange}
                            className={styles.createInsightRadio}/>
                        <FormRadioInput
                            title='Years'
                            name='step'
                            value='years'
                            checked={step.input.value === 'years'}
                            onChange={step.input.onChange}
                            className={styles.createInsightRadio}/>
                    </div>
                </FormGroup>

                <div className={styles.createInsightButtons}>
                    <button
                        type='submit'
                        className={classnames(styles.createInsightButton, styles.createInsightButtonActive, 'button')}>

                        Create code insight
                    </button>
                    <button
                        type='button'
                        className={classnames(styles.createInsightButton, 'button')}>

                        Cancel
                    </button>
                </div>
            </form>
        </Page>
    )
}
