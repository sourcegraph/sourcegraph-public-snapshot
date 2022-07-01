import { FunctionComponent, HTMLAttributes } from 'react'

import { CreationUiLayout, CreationUIForm, CreationUIPreview } from '../../../../../components'
import { FormChangeEvent, SubmissionErrors } from '../../../../../components/form/hooks/useForm'
import { ComputeLivePreview } from '../../ComputeLivePreview'
import { EditableDataSeries } from '../../search-insight'
import { useEditableSeries } from '../../search-insight/components/search-insight-creation-content/hooks/use-editable-series'
import { useInsightCreationForm } from '../../search-insight/components/search-insight-creation-content/hooks/use-insight-creation-form/use-insight-creation-form'
import { getSanitizedSeries } from '../../search-insight/utils/insight-sanitizer'

import { CreateComputeInsightFormFields } from './types'

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

    const { repositories, series, step, stepValue, allReposMode } = useInsightCreationForm({
        mode,
        initialValue,
        onChange,
        onSubmit,
    })

    const { editSeries } = useEditableSeries({ series })

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
            <CreationUIForm>Hello World</CreationUIForm>

            <CreationUIPreview
                as={ComputeLivePreview}
                disabled={!allFieldsForPreviewAreValid}
                repositories={repositories.meta.value}
                isAllReposMode={allReposMode.input.value}
                series={seriesToPreview(editSeries)}
                step={step.meta.value}
                stepValue={stepValue.meta.value}
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
