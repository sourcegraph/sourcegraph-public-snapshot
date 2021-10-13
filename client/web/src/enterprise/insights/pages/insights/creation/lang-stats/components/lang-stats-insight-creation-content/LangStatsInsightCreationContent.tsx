import classnames from 'classnames'
import React, { useCallback, useContext } from 'react'
import { noop } from 'rxjs'

import { useField } from '../../../../../../components/form/hooks/useField'
import { FormChangeEvent, SubmissionErrors, useForm } from '../../../../../../components/form/hooks/useForm'
import { AsyncValidator } from '../../../../../../components/form/hooks/utils/use-async-validation'
import { createRequiredValidator } from '../../../../../../components/form/validators'
import { InsightsApiContext } from '../../../../../../core/backend/api-provider'
import { InsightTypePrefix } from '../../../../../../core/types'
import { isUserSubject, SupportedInsightSubject } from '../../../../../../core/types/subjects'
import { LangStatsCreationFormFields } from '../../types'
import { LangStatsInsightCreationForm } from '../lang-stats-insight-creation-form/LangStatsInsightCreationForm'
import { LangStatsInsightLivePreview } from '../live-preview-chart/LangStatsInsightLivePreview'

import styles from './LangStatsInsightCreationContent.module.scss'
import { repositoriesFieldValidator, repositoryFieldAsyncValidator, thresholdFieldValidator } from './validators'

const INITIAL_VALUES: LangStatsCreationFormFields = {
    repository: '',
    title: '',
    threshold: 3,
    visibility: 'personal',
}

const titleRequiredValidator = createRequiredValidator('Title is a required field.')

export interface LangStatsInsightCreationContentProps {
    /**
     * This component might be used in two different modes for creation and
     * edit mode. In edit mode we change some text keys for form and trigger
     * validation on form fields immediately.
     */
    mode?: 'creation' | 'edit'

    subjects?: SupportedInsightSubject[]

    /** Initial value for all form fields. */
    initialValues?: Partial<LangStatsCreationFormFields>

    /** Custom class name for root form element. */
    className?: string

    /** Submit handler for form element. */
    onSubmit: (values: LangStatsCreationFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void

    /** Change handlers is called every time when user changed any field within the form. */
    onChange?: (event: FormChangeEvent<LangStatsCreationFormFields>) => void

    /** Cancel handler. */
    onCancel?: () => void
}

export const LangStatsInsightCreationContent: React.FunctionComponent<LangStatsInsightCreationContentProps> = props => {
    const {
        mode = 'creation',
        subjects = [],
        initialValues = {},
        className,
        onSubmit,
        onCancel = noop,
        onChange = noop,
    } = props

    const { findInsightByName } = useContext(InsightsApiContext)

    const { values, handleSubmit, formAPI, ref } = useForm<LangStatsCreationFormFields>({
        initialValues: {
            ...INITIAL_VALUES,
            // Calculate initial value for the visibility settings
            visibility: subjects.find(isUserSubject)?.id ?? '',
            ...initialValues,
        },
        onSubmit,
        onChange,
        touched: mode === 'edit',
    })

    const repository = useField({
        name: 'repository',
        formApi: formAPI,
        validators: {
            sync: repositoriesFieldValidator,
            async: repositoryFieldAsyncValidator,
        },
    })

    const asyncTitleValidator = useCallback<AsyncValidator<string>>(
        async title => {
            if (!title || title.trim() === '' || title === initialValues?.title) {
                return
            }

            const possibleInsight = await findInsightByName(title, InsightTypePrefix.langStats).toPromise()

            if (possibleInsight) {
                return 'An insight with this name already exists. Please set a different name for the new insight.'
            }

            return
        },
        [findInsightByName, initialValues?.title]
    )

    const title = useField({
        name: 'title',
        formApi: formAPI,
        validators: { sync: titleRequiredValidator, async: asyncTitleValidator },
    })

    const threshold = useField({
        name: 'threshold',
        formApi: formAPI,
        validators: { sync: thresholdFieldValidator },
    })
    const visibility = useField({
        name: 'visibility',
        formApi: formAPI,
    })

    // If some fields that needed to run live preview  are invalid
    // we should disabled live chart preview
    const allFieldsForPreviewAreValid =
        repository.meta.validState === 'VALID' ||
        (repository.meta.validState === 'CHECKING' && threshold.meta.validState === 'VALID')

    const handleFormReset = (): void => {
        // TODO [VK] Change useForm API in order to implement form.reset method.
        title.input.onChange('')
        repository.input.onChange('')
        // Focus first element of the form
        repository.input.ref.current?.focus()
        visibility.input.onChange('personal')
        threshold.input.onChange(3)
    }

    const hasFilledValue = values.repository !== '' || values.title !== ''

    return (
        <div data-testid="code-stats-insight-creation-page-content" className={classnames(styles.content, className)}>
            <LangStatsInsightCreationForm
                mode={mode}
                innerRef={ref}
                handleSubmit={handleSubmit}
                submitErrors={formAPI.submitErrors}
                submitting={formAPI.submitting}
                title={title}
                repository={repository}
                threshold={threshold}
                visibility={visibility}
                subjects={subjects}
                isFormClearActive={hasFilledValue}
                onCancel={onCancel}
                className={styles.contentForm}
                onFormReset={handleFormReset}
            />

            <LangStatsInsightLivePreview
                repository={repository.meta.value}
                threshold={threshold.meta.value}
                disabled={!allFieldsForPreviewAreValid}
                className={styles.contentLivePreview}
            />
        </div>
    )
}
