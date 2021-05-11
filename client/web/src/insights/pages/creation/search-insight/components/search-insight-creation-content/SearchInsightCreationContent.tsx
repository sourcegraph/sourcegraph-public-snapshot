import classnames from 'classnames';
import React from 'react';
import { noop } from 'rxjs';

import { Settings } from '@sourcegraph/shared/src/settings/settings';

import { useField, Validator } from '../../../../../components/form/hooks/useField';
import { SubmissionErrors, useForm } from '../../../../../components/form/hooks/useForm';
import { useTitleValidator } from '../../../../../components/form/hooks/useTitleValidator';
import { createRequiredValidator } from '../../../../../components/form/validators';
import { InsightTypeSuffix } from '../../../../../core/types';
import { CreateInsightFormFields, DataSeries } from '../../types';
import { SearchInsightLivePreview } from '../live-preview-chart/SearchInsightLivePreview';
import { SearchInsightCreationForm } from '../search-insight-creation-form/SearchInsightCreationForm';

import styles from './SearchInsightCreationContent.module.scss'

const repositoriesFieldValidator = createRequiredValidator('Repositories is a required field.')
const requiredStepValueField = createRequiredValidator('Please specify a step between points.')
/**
 * Custom validator for chart series. Since series has complex type
 * we can't validate this with standard validators.
 * */
const seriesRequired: Validator<DataSeries[]> = series =>
    series && series.length > 0 ? undefined : 'Series is empty. You must have at least one series for code insight.'

const INITIAL_VALUES: CreateInsightFormFields = {
    visibility: 'personal',
    series: [],
    step: 'months',
    stepValue: '2',
    title: '',
    repositories: '',
}

export interface SearchInsightCreationContentProps {
    /** Final settings cascade. Used for title field validation. */
    settings?: Settings | null
    /** Initial value for all form fields. */
    initialValue?: CreateInsightFormFields
    /** Custom class name for root form element. */
    className?: string
    /** Submit handler for form element. */
    onSubmit: (values: CreateInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    /** Cancel handler. */
    onCancel?: () => void
}

export const SearchInsightCreationContent: React.FunctionComponent<SearchInsightCreationContentProps> = props => {
    const { settings, initialValue = INITIAL_VALUES, onSubmit, onCancel = noop, className } = props;

    const { formAPI, ref, handleSubmit } = useForm<CreateInsightFormFields>({
        initialValues: initialValue,
        onSubmit,
        onChange: values => {
            console.log('next values', values);
        }
    })

    // We can't have two or more insights with the same name, since we rely on name as on id of insights.
    const titleValidator = useTitleValidator({ settings, insightType: InsightTypeSuffix.search })

    const title = useField('title', formAPI, titleValidator)
    const repositories = useField('repositories', formAPI, repositoriesFieldValidator)
    const visibility = useField('visibility', formAPI)

    const series = useField('series', formAPI, seriesRequired)
    const step = useField('step', formAPI)
    const stepValue = useField('stepValue', formAPI, requiredStepValueField)

    // If some fields that needed to run live preview  are invalid
    // we should disabled live chart preview
    const allFieldsForPreviewAreValid =
        repositories.meta.validState === 'VALID' &&
        series.meta.validState === 'VALID' &&
        stepValue.meta.validState === 'VALID'

    return (
        <div className={classnames(styles.content, className)}>
            <SearchInsightCreationForm
                innerRef={ref}
                handleSubmit={handleSubmit}
                submitErrors={formAPI.submitErrors}
                submitting={formAPI.submitting}
                title={title}
                repositories={repositories}
                visibility={visibility}
                series={series}
                step={step}
                stepValue={stepValue}
                onCancel={onCancel}
            />

            <SearchInsightLivePreview
                disabled={!allFieldsForPreviewAreValid}
                repositories={repositories.meta.value}
                series={series.meta.value}
                step={step.meta.value}
                stepValue={stepValue.meta.value}
                className={classnames(styles.contentLivePreview)}/>
        </div>
    );
}
