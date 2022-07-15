import { FC, HTMLAttributes } from 'react'

import { Code, Input, Link } from '@sourcegraph/wildcard'

import {
    createDefaultEditSeries,
    CreationUIForm,
    CreationUiLayout,
    CreationUIPreview,
    FormChangeEvent,
    FormGroup,
    FormSeries,
    getDefaultInputProps,
    insightRepositoriesAsyncValidator,
    insightRepositoriesValidator,
    insightTitleValidator,
    RepositoriesField,
    SubmissionErrors,
    useField,
    EditableDataSeries,
    useForm,
} from '../../../../../components'
import { useEditableSeries } from '../../../../../components/creation-ui/form-series/use-editable-series'
import { useUiFeatures } from '../../../../../hooks'
import { ComputeLivePreview } from '../../ComputeLivePreview'
import { getSanitizedSeries } from '../../search-insight/utils/insight-sanitizer'
import { ComputeInsightMap, CreateComputeInsightFormFields } from '../types'

import { ComputeInsightMapPicker } from './ComputeInsightMapPicker'

const INITIAL_INSIGHT_VALUES: CreateComputeInsightFormFields = {
    series: [createDefaultEditSeries({ edit: true })],
    title: '',
    repositories: '',
    groupBy: ComputeInsightMap.Repositories,
    dashboardReferenceCount: 0,
}

type NativeContainerProps = Omit<HTMLAttributes<HTMLDivElement>, 'onSubmit' | 'onChange'>

interface ComputeInsightCreationContentProps extends NativeContainerProps {
    /** This component might be used in edit or creation insight case. */
    mode?: 'creation' | 'edit'
    initialValue?: Partial<CreateComputeInsightFormFields>

    onChange: (event: FormChangeEvent<CreateComputeInsightFormFields>) => void
    onSubmit: (values: CreateComputeInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const ComputeInsightCreationContent: FC<ComputeInsightCreationContentProps> = props => {
    const { mode = 'creation', initialValue, onChange, onSubmit, onCancel, ...attributes } = props

    const { licensed } = useUiFeatures()

    const { formAPI, handleSubmit } = useForm<CreateComputeInsightFormFields>({
        initialValues: { ...INITIAL_INSIGHT_VALUES, ...initialValue },
        onSubmit,
        onChange,
        touched: mode === 'edit',
    })

    const title = useField({
        name: 'title',
        formApi: formAPI,
        validators: { sync: insightTitleValidator },
    })

    const repositories = useField({
        name: 'repositories',
        formApi: formAPI,
        validators: {
            // Turn off any validations for the repositories' field in we are in all repos mode
            sync: insightRepositoriesValidator,
            async: insightRepositoriesAsyncValidator,
        },
    })

    const series = useField({
        name: 'series',
        formApi: formAPI,
    })

    const groupBy = useField({
        name: 'groupBy',
        formApi: formAPI,
    })

    const { series: editSeries } = useEditableSeries(series)

    // If some fields that needed to run live preview  are invalid
    // we should disable live chart preview
    const allFieldsForPreviewAreValid =
        repositories.meta.validState === 'VALID' &&
        (series.meta.validState === 'VALID' || editSeries.some(series => series.valid))

    return (
        <CreationUiLayout {...attributes}>
            <CreationUIForm onSubmit={handleSubmit}>
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

                <hr className="my-4 w-100" />

                <FormGroup
                    innerRef={series.input.ref}
                    name="data series group"
                    title="Data series"
                    subtitle={
                        licensed
                            ? 'Add any number of data series to your chart'
                            : 'Add up to 10 data series to your chart'
                    }
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

                <hr className="my-4 w-100" />

                <FormGroup name="map result" title="Map result">
                    <ComputeInsightMapPicker series={series.input.value} {...groupBy.input} />

                    <small className="text-muted mt-3">
                        Learn more about <Link to="">grouping results</Link>
                    </small>
                </FormGroup>

                <hr className="my-4 w-100" />

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

            <CreationUIPreview
                as={ComputeLivePreview}
                disabled={!allFieldsForPreviewAreValid}
                repositories={repositories.meta.value}
                series={seriesToPreview(editSeries)}
            />
        </CreationUiLayout>
    )
}

function seriesToPreview(
    currentSeries: EditableDataSeries[]
): {
    query: string
    label: string
    generatedFromCaptureGroup: boolean
    stroke: string
}[] {
    const validSeries = currentSeries.filter(series => series.valid)
    return getSanitizedSeries(validSeries).map(series => ({
        query: series.query,
        stroke: series.stroke ? series.stroke : '',
        label: series.name,
        generatedFromCaptureGroup: false,
    }))
}
