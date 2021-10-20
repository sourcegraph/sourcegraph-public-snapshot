import { useAsyncInsightTitleValidator } from '../../../../../../../../components/form/hooks/use-async-insight-title-validator'
import { useField, useFieldAPI } from '../../../../../../../../components/form/hooks/useField'
import { Form, FormChangeEvent, SubmissionErrors, useForm } from '../../../../../../../../components/form/hooks/useForm'
import { createRequiredValidator } from '../../../../../../../../components/form/validators'
import { isUserSubject, SupportedInsightSubject } from '../../../../../../../../core/types/subjects'
import { CreateInsightFormFields, EditableDataSeries, InsightStep } from '../../../../types'
import { INITIAL_INSIGHT_VALUES } from '../../initial-insight-values'
import {
    repositoriesExistValidator,
    repositoriesFieldValidator,
    requiredStepValueField,
    seriesRequired,
} from '../../validators'

const titleRequiredValidator = createRequiredValidator('Title is a required field.')

export interface UseInsightCreationFormProps {
    mode: 'creation' | 'edit'
    subjects?: SupportedInsightSubject[]
    initialValue?: Partial<CreateInsightFormFields>
    onSubmit: (values: CreateInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onChange?: (event: FormChangeEvent<CreateInsightFormFields>) => void
}

export interface InsightCreationForm {
    form: Form<CreateInsightFormFields>
    title: useFieldAPI<string>
    repositories: useFieldAPI<string>
    visibility: useFieldAPI<string>
    series: useFieldAPI<EditableDataSeries[]>
    step: useFieldAPI<InsightStep>
    stepValue: useFieldAPI<string>
    allReposMode: useFieldAPI<boolean>
}

/**
 * Hooks absorbs all insight creation form logic (field state managements,
 * validations, fields dependencies)
 */
export function useInsightCreationForm(props: UseInsightCreationFormProps): InsightCreationForm {
    const { mode, subjects = [], initialValue = {}, onSubmit, onChange } = props

    // Calculate initial value for visibility settings
    const userSubjectID = subjects.find(isUserSubject)?.id ?? ''
    const isEdit = mode === 'edit'

    const form = useForm<CreateInsightFormFields>({
        initialValues: {
            ...INITIAL_INSIGHT_VALUES,
            visibility: userSubjectID,
            ...initialValue,
        },
        onSubmit,
        onChange,
        touched: isEdit,
    })

    const allReposMode = useField({
        name: 'allRepos',
        formApi: form.formAPI,
        onChange: (checked: boolean) => {
            // Reset form values in case if All repos mode was activated
            if (checked) {
                repositories.input.onChange('')
                step.input.onChange('weeks')
                stepValue.input.onChange('2')
            }
        },
    })

    const isAllReposMode = allReposMode.input.value
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
            // Turn off any validations for the repositories field in we are in all repos mode
            sync: !isAllReposMode ? repositoriesFieldValidator : undefined,
            async: !isAllReposMode ? repositoriesExistValidator : undefined,
        },
        disabled: isAllReposMode,
    })

    const visibility = useField({
        name: 'visibility',
        formApi: form.formAPI,
    })

    const series = useField({
        name: 'series',
        formApi: form.formAPI,
        validators: { sync: seriesRequired },
    })

    const step = useField({
        name: 'step',
        formApi: form.formAPI,
        disabled: isAllReposMode,
    })
    const stepValue = useField({
        name: 'stepValue',
        formApi: form.formAPI,
        validators: {
            // Turn off any validations if we are in all repos mode
            sync: !isAllReposMode ? requiredStepValueField : undefined,
        },
        disabled: isAllReposMode,
    })

    return {
        form,
        title,
        repositories,
        visibility,
        series,
        step,
        stepValue,
        allReposMode,
    }
}
