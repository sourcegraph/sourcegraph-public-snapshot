import classnames from 'classnames';
import React  from 'react';
import { useField, useForm } from 'react-final-form-hooks';

import { Page } from '../../../components/Page';
import { PageTitle } from '../../../components/PageTitle';

import { InputField } from './components/form-field/FormField';
import { FormGroup } from './components/form-group/FormGroup';
import { FormRadioInput } from './components/form-radio-input/FormRadioInput';
import { FormSeries } from './components/form-series/FormSeries';
import styles from './CreateinsightPage.module.scss'
import { DataSeries } from './types';

interface CreateInsightPageProps {}

interface Form {
    series: DataSeries[];
    title: string;
    repositories: string;
    visibility: string;
}

const seriesRequired = (series: DataSeries[]) => series && series.length > 0 ? undefined : 'Required';

export const CreateInsightPage: React.FunctionComponent<CreateInsightPageProps> = props => {
    const {} = props;

    const form = useForm<Form>({
        validate: values => {
            return { series: seriesRequired(values.series)}
        },
        onSubmit: () => {
            console.log('Submit');
        }
    });

    const title = useField('title', form.form);
    const repositories = useField('repositories', form.form);
    const visibility = useField('visibility', form.form);
    const series = useField<DataSeries[], Form>('series', form.form, seriesRequired);
    const step = useField('step', form.form);

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
            <form onSubmit={form.handleSubmit} className={styles.createInsightForm}>

                <InputField
                    title='Title'
                    description='Shown as title for your insight'
                    placeholder='ex. Migration to React function components'
                    error={title.meta.touched && title.meta.error}
                    {...title.input}
                    className={styles.createInsightFormField}/>

                <InputField
                    title='Repositories'
                    description='Create a list of repositories to run your search over. Separate them with comas.'
                    placeholder='Add or search for repositories'
                    error={repositories.meta.touched && repositories.meta.error}
                    {...repositories.input}
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

                <hr className={styles.createInsightSeparator}/>

                <FormGroup
                    name='Data series'
                    subtitle='Add any number of data series to your chart'>

                    <FormSeries
                        series={series.input.value}
                        onChange={series.input.onChange}/>

                    { series.meta.touched && series.meta.error && <div>{series.meta.error}</div> }

                </FormGroup>

                <hr className={styles.createInsightSeparator}/>

                <FormGroup
                    name='Step between data points'
                    description='The distance between two data points on the chart'
                    className={styles.createInsightFormField}>

                    <div className={styles.createInsightRadioGroupContent}>

                        <input
                            type="text"
                            placeholder='ex. 2'
                            className={classnames(styles.createInsightStepInput, 'form-control')}
                        />

                        <FormRadioInput
                            title='Hours'
                            name='step'
                            className={styles.createInsightRadio}/>
                        <FormRadioInput
                            title='Days'
                            name='step'
                            className={styles.createInsightRadio}/>
                        <FormRadioInput
                            title='Weeks'
                            name='step'
                            className={styles.createInsightRadio}/>
                        <FormRadioInput
                            title='Months'
                            name='step'
                            className={styles.createInsightRadio}/>
                        <FormRadioInput
                            title='Years'
                            name='step'
                            className={styles.createInsightRadio}/>
                    </div>
                </FormGroup>

                <hr className={styles.createInsightSeparator}/>

                <div>
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

                <pre>{JSON.stringify(form.values, 0, 2)}</pre>
                <pre>{JSON.stringify(form.errors, 0, 2)}</pre>
            </form>
        </Page>
    )
}
