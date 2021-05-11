import classnames from 'classnames';
import React from 'react'
import { noop } from 'rxjs';

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { useField } from '../../../../../components/form/hooks/useField';
import { SubmissionErrors, useForm } from '../../../../../components/form/hooks/useForm';
import { useTitleValidator } from '../../../../../components/form/hooks/useTitleValidator';
import { createRequiredValidator } from '../../../../../components/form/validators';
import { InsightTypeSuffix } from '../../../../../core/types';
import { LangStatsCreationFormFields } from '../../types';
import { LangStatsInsightCreationForm } from '../lang-stats-insight-creation-form/LangStatsInsightCreationForm';

import styles from './LangStatsInsightCreationContent.module.scss'

const repositoriesFieldValidator = createRequiredValidator('Repositories is a required field for code insight.')
const thresholdFieldValidator = createRequiredValidator('Threshold is a required field for code insight.')

const INITIAL_VALUES: LangStatsCreationFormFields = {
    repository: '',
    title: '',
    threshold: 3,
    visibility: 'personal',
}

export interface LangStatsInsightCreationContentProps {
    /** Final settings cascade. Used for title field validation. */
    settings?: Settings | null
    /** Initial value for all form fields. */
    initialValues?: LangStatsCreationFormFields
    /** Custom class name for root form element. */
    className?: string
    /** Submit handler for form element. */
    onSubmit: (values: LangStatsCreationFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    /** Cancel handler. */
    onCancel?: () => void
}

export const LangStatsInsightCreationContent: React.FunctionComponent<LangStatsInsightCreationContentProps> = props => {
    const { settings, initialValues = INITIAL_VALUES, className, onSubmit, onCancel = noop } = props;

    const { handleSubmit, formAPI, ref } = useForm<LangStatsCreationFormFields>({
        initialValues,
        onSubmit,
    })

    // We can't have two or more insights with the same name, since we rely on name as on id of insights.
    const titleValidator = useTitleValidator({ settings, insightType: InsightTypeSuffix.langStats })

    const repository = useField('repository', formAPI, repositoriesFieldValidator)
    const title = useField('title', formAPI, titleValidator)
    const threshold = useField('threshold', formAPI, thresholdFieldValidator)
    const visibility = useField('visibility', formAPI)

    return (
        <div className={classnames(styles.content, className)}>

            <LangStatsInsightCreationForm
                innerRef={ref}
                handleSubmit={handleSubmit}
                submitErrors={formAPI.submitErrors}
                submitting={formAPI.submitting}
                title={title}
                repository={repository}
                threshold={threshold}
                visibility={visibility}
                onCancel={onCancel}
                className={styles.contentForm}
            />
        </div>
    );
}
