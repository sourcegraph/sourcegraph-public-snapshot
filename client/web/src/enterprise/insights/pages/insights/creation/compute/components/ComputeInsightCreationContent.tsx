import { FC, HTMLAttributes } from 'react'

import { Code, Input } from '@sourcegraph/wildcard'

import {
    CreationUiLayout,
    CreationUIForm,
    CreationUIPreview,
    FormChangeEvent,
    SubmissionErrors,
    FormSeries,
    FormGroup,
    useForm,
    createDefaultEditSeries,
    useField,
    insightTitleValidator,
    insightRepositoriesValidator,
    insightRepositoriesAsyncValidator,
    RepositoriesField,
    getDefaultInputProps,
} from '../../../../../components'
import { useUiFeatures } from '../../../../../hooks'

import { CreateComputeInsightFormFields } from './types'

const INITIAL_INSIGHT_VALUES: CreateComputeInsightFormFields = {
    series: [createDefaultEditSeries({ edit: true })],
    title: '',
    repositories: '',
    dashboardReferenceCount: 0,
}

type NativeContainerProps = Omit<HTMLAttributes<HTMLDivElement>, 'onSubmit' | 'onChange'>

interface ComputeInsightCreationContentProps extends NativeContainerProps {
    mode?: 'creation' | 'edit'
    initialValue?: Partial<CreateComputeInsightFormFields>

    onChange: (event: FormChangeEvent<CreateComputeInsightFormFields>) => void
    onSubmit: (values: CreateComputeInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const ComputeInsightCreationContent: FC<ComputeInsightCreationContentProps> = props => {
    const { mode = 'creation', initialValue, onChange, onSubmit, onCancel, ...attributes } = props
    const { licensed } = useUiFeatures()

    const form = useForm<CreateComputeInsightFormFields>({
        initialValues: { ...INITIAL_INSIGHT_VALUES, ...initialValue },
        onSubmit,
        onChange,
        touched: mode === 'edit',
    })

    const title = useField({
        name: 'title',
        formApi: form.formAPI,
        validators: { sync: insightTitleValidator },
    })

    const repositories = useField({
        name: 'repositories',
        formApi: form.formAPI,
        validators: {
            // Turn off any validations for the repositories field in we are in all repos mode
            sync: insightRepositoriesValidator,
            async: insightRepositoriesAsyncValidator,
        },
    })

    const series = useField({
        name: 'series',
        formApi: form.formAPI,
    })

    return (
        <CreationUiLayout {...attributes}>
            <CreationUIForm>
                <FormGroup
                    name="insight repositories"
                    title="Targeted repositories"
                    subtitle="Create a list of repositories to run your search over"
                >
                    <Input
                        as={RepositoriesField}
                        autoFocus={true}
                        required={true}
                        label="Repositories"
                        message="Separate repositories with commas"
                        placeholder="Example: github.com/sourcegraph/sourcegraph"
                        {...getDefaultInputProps(repositories)}
                        className="mb-0 d-flex flex-column"
                    />
                </FormGroup>

                <FormGroup
                    name="data series group"
                    title="Data series"
                    subtitle={
                        licensed
                            ? 'Add any number of data series to your chart'
                            : 'Add up to 10 data series to your chart'
                    }
                    innerRef={series.input.ref}
                >
                    <FormSeries
                        seriesField={series}
                        repositories={repositories.input.value}
                        showValidationErrorsOnMount={false}
                        queryFieldDescription={
                            <ul className="pl-3">
                                <li>
                                    Do not include the <Code weight="bold">repo:</Code> filter as it will be added
                                    automatically, if needed{' '}
                                </li>
                                <li>
                                    You can use <Code weight="bold">before:</Code> and <Code weight="bold">after:</Code>{' '}
                                    operators for <Code weight="bold">type:diff</Code> and{' '}
                                    <Code weight="bold">type:commit</Code> to define the timeframe (example query:{' '}
                                    <Code>type:diff author:nick before:"last thursday" SearchTerm</Code>)
                                </li>
                            </ul>
                        }
                    />
                </FormGroup>

                <FormGroup name="chart settings group" title="Chart settings">
                    <Input
                        label="Title"
                        required={true}
                        message="Shown as the title for your insight"
                        placeholder="Example: Migration to React function components"
                        className="d-flex flex-column"
                        {...getDefaultInputProps(title)}
                    />
                </FormGroup>
            </CreationUIForm>

            <CreationUIPreview>This is live preview</CreationUIPreview>
        </CreationUiLayout>
    )
}
