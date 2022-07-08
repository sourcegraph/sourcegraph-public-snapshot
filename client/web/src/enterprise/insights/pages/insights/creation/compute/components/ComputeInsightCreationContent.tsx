import { FunctionComponent, HTMLAttributes } from 'react'

import { Code } from '@sourcegraph/wildcard'

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
    EditableDataSeries,
} from '../../../../../components'
import { useEditableSeries } from '../../../../../components/creation-ui/form-series/use-editable-series'
import { useUiFeatures } from '../../../../../hooks'
import { ComputeLivePreview } from '../../ComputeLivePreview'
import { useInsightCreationForm } from '../../search-insight/components/search-insight-creation-content/hooks/use-insight-creation-form'
import { getSanitizedSeries } from '../../search-insight/utils/insight-sanitizer'

import { CreateComputeInsightFormFields } from './types'

const INITIAL_INSIGHT_VALUES: CreateComputeInsightFormFields = {
    series: [createDefaultEditSeries({ edit: true })],
    title: '',
    repositories: '',
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

export const ComputeInsightCreationContent: FunctionComponent<ComputeInsightCreationContentProps> = props => {
    const { mode = 'creation', initialValue, onChange, onSubmit, onCancel, ...attributes } = props

    const { repositories, stepValue, allReposMode } = useInsightCreationForm({
        mode,
        initialValue,
        onChange,
        onSubmit,
    })

    const { licensed } = useUiFeatures()

    const form = useForm<CreateComputeInsightFormFields>({
        initialValues: { ...INITIAL_INSIGHT_VALUES, ...initialValue },
        onSubmit,
        onChange,
        touched: mode === 'edit',
    })

    const series = useField({
        name: 'series',
        formApi: form.formAPI,
    })

    const { series: editSeries } = useEditableSeries(series)

    // If some fields that needed to run live preview  are invalid
    // we should disable live chart preview
    const allFieldsForPreviewAreValid =
        repositories.meta.validState === 'VALID' &&
        (series.meta.validState === 'VALID' || editSeries.some(series => series.valid)) &&
        stepValue.meta.validState === 'VALID' &&
        // For the "all repositories" mode we are not able to show the live preview chart
        !allReposMode.input.value

    return (
        <CreationUiLayout {...attributes}>
            <CreationUIForm>
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
                        repositories=""
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
