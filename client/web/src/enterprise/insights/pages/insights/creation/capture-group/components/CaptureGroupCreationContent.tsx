import classNames from 'classnames'
import { noop } from 'lodash'
import React, { useCallback } from 'react'

import styles from '../../../../../components/creation-ui-kit/CreationUiKit.module.scss'
import { useAsyncInsightTitleValidator } from '../../../../../components/form/hooks/use-async-insight-title-validator'
import { useField } from '../../../../../components/form/hooks/useField'
import { FormChangeEvent, SubmissionErrors, useForm } from '../../../../../components/form/hooks/useForm'
import { createRequiredValidator } from '../../../../../components/form/validators'
import {
    repositoriesExistValidator,
    repositoriesFieldValidator,
    requiredStepValueField,
} from '../../search-insight/components/search-insight-creation-content/validators'
import { searchQueryValidator } from '../search-query-validator'
import { CaptureGroupFormFields } from '../types'

import { CaptureGroupCreationForm } from './CaptureGoupCreationForm'
import { CaptureGroupCreationLivePreview } from './CaptureGroupCreationLivePreview'

const INITIAL_VALUES: CaptureGroupFormFields = {
    repositories: '',
    groupSearchQuery: '',
    title: '',
    step: 'months',
    stepValue: '2',
}

const titleRequiredValidator = createRequiredValidator('Title is a required field.')
const queryRequiredValidator = createRequiredValidator('Query is a required field.')

interface CaptureGroupCreationContentProps {
    mode: 'creation' | 'edit'
    initialValues?: Partial<CaptureGroupFormFields>
    className?: string

    onSubmit: (values: CaptureGroupFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onChange?: (event: FormChangeEvent<CaptureGroupFormFields>) => void
    onCancel: () => void
}

export const CaptureGroupCreationContent: React.FunctionComponent<CaptureGroupCreationContentProps> = props => {
    const { mode, className, initialValues = {}, onSubmit, onChange = noop, onCancel } = props

    // Search query validators
    const validateChecks = useCallback((value: string | undefined) => {
        if (!value) {
            return queryRequiredValidator(value)
        }
        const validatedChecks = searchQueryValidator(value, value !== undefined)
        const allChecksPassed = Object.values(validatedChecks).every(Boolean)

        if (!allChecksPassed) {
            return 'Query is not valid'
        }

        return queryRequiredValidator(value)
    }, [])

    const form = useForm<CaptureGroupFormFields>({
        initialValues: { ...INITIAL_VALUES, ...initialValues },
        touched: mode === 'edit',
        onSubmit,
        onChange,
    })

    const asyncTitleValidator = useAsyncInsightTitleValidator({
        mode,
        initialTitle: form.formAPI.initialValues.title,
    })

    const title = useField({
        name: 'title',
        formApi: form.formAPI,
        validators: { sync: titleRequiredValidator, async: asyncTitleValidator },
    })

    const repositories = useField({
        name: 'repositories',
        formApi: form.formAPI,
        validators: {
            sync: repositoriesFieldValidator,
            async: repositoriesExistValidator,
        },
    })

    const query = useField({
        name: 'groupSearchQuery',
        formApi: form.formAPI,
        validators: { sync: validateChecks },
    })

    const step = useField({
        name: 'step',
        formApi: form.formAPI,
    })

    const stepValue = useField({
        name: 'stepValue',
        formApi: form.formAPI,
        validators: { sync: requiredStepValueField },
    })

    const handleFormReset = (): void => {
        title.input.onChange('')
        repositories.input.onChange('')
        query.input.onChange('')
        step.input.onChange('months')
        stepValue.input.onChange('1')

        // Focus first element of the form
        repositories.input.ref.current?.focus()
    }

    const hasFilledValue =
        form.values.title !== '' || form.values.repositories !== '' || form.values.groupSearchQuery !== ''

    const areAllFieldsForPreviewValid =
        repositories.meta.validState === 'VALID' &&
        stepValue.meta.validState === 'VALID' &&
        query.meta.validState === 'VALID'

    return (
        <div className={classNames(styles.content, className)}>
            <CaptureGroupCreationForm
                mode={mode}
                form={form}
                title={title}
                repositories={repositories}
                step={step}
                stepValue={stepValue}
                query={query}
                isFormClearActive={hasFilledValue}
                onCancel={onCancel}
                onFormReset={handleFormReset}
                className={styles.contentForm}
            />

            <CaptureGroupCreationLivePreview
                disabled={!areAllFieldsForPreviewValid}
                repositories={repositories.meta.value}
                query={query.meta.value}
                step={step.meta.value}
                stepValue={stepValue.meta.value}
                className={styles.contentLivePreview}
            />
        </div>
    )
}
