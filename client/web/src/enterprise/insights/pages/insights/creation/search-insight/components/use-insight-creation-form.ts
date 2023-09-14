import type { QueryState } from '@sourcegraph/shared/src/search'
import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import {
    type FormInstance,
    type FormChangeEvent,
    type SubmissionErrors,
    useField,
    type useFieldAPI,
    useForm,
} from '@sourcegraph/wildcard'

import {
    createDefaultEditSeries,
    type EditableDataSeries,
    insightSeriesValidator,
    insightStepValueValidator,
    insightTitleValidator,
    useRepoFields,
} from '../../../../../components'
import type { CreateInsightFormFields, InsightStep, RepoMode } from '../types'

export const INITIAL_INSIGHT_VALUES: CreateInsightFormFields = {
    // If user opens the creation form to create insight
    // we want to show the series form as soon as possible
    // and do not force the user to click the 'add another series' button
    series: [createDefaultEditSeries({ edit: true })],
    step: 'months',
    stepValue: '2',
    title: '',
    repositories: [],
    repoMode: 'search-query',
    repoQuery: { query: '' },
    dashboardReferenceCount: 0,
}

export interface UseInsightCreationFormProps {
    touched: boolean
    initialValue?: Partial<CreateInsightFormFields>
    onSubmit: (values: CreateInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onChange?: (event: FormChangeEvent<CreateInsightFormFields>) => void
}

export interface InsightCreationForm {
    form: FormInstance<CreateInsightFormFields>
    title: useFieldAPI<string>
    repositories: useFieldAPI<string[]>
    repoQuery: useFieldAPI<QueryState>
    repoMode: useFieldAPI<RepoMode>
    series: useFieldAPI<EditableDataSeries[]>
    step: useFieldAPI<InsightStep>
    stepValue: useFieldAPI<string>
}

/**
 * Hooks absorbs all insight creation form logic (field state managements,
 * validations, fields dependencies)
 */
export function useInsightCreationForm(props: UseInsightCreationFormProps): InsightCreationForm {
    const { touched, initialValue = {}, onSubmit, onChange } = props

    const repoFieldVariation = useExperimentalFeatures(features => features.codeInsightsRepoUI)
    const isSearchQueryORUrlsList = repoFieldVariation === 'search-query-or-strict-list'

    // Enforce "search-query" initial value if we're in the single search query UI mode
    const initialValues = isSearchQueryORUrlsList
        ? { ...INITIAL_INSIGHT_VALUES, ...initialValue }
        : { ...INITIAL_INSIGHT_VALUES, ...initialValue, repoMode: 'search-query' as const }

    const form = useForm<CreateInsightFormFields>({
        initialValues,
        touched,
        onSubmit,
        onChange,
    })

    const { repoMode, repoQuery, repositories } = useRepoFields({ formApi: form.formAPI })

    const title = useField({
        name: 'title',
        formApi: form.formAPI,
        validators: { sync: insightTitleValidator },
    })

    const series = useField({
        name: 'series',
        formApi: form.formAPI,
        validators: { sync: insightSeriesValidator },
    })

    const step = useField({
        name: 'step',
        formApi: form.formAPI,
    })

    const stepValue = useField({
        name: 'stepValue',
        formApi: form.formAPI,
        validators: { sync: insightStepValueValidator },
    })

    return {
        form,
        title,
        repositories,
        repoQuery,
        repoMode,
        series,
        step,
        stepValue,
    }
}
