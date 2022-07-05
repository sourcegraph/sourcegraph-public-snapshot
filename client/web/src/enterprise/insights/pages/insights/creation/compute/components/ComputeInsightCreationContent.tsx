import { FunctionComponent, HTMLAttributes } from 'react'

import { CreationUiLayout, CreationUIForm, CreationUIPreview } from '../../../../../components'
import { FormChangeEvent, SubmissionErrors } from '../../../../../components/form/hooks/useForm'

import { CreateComputeInsightFormFields } from './types'

type NativeContainerProps = Omit<HTMLAttributes<HTMLDivElement>, 'onSubmit' | 'onChange'>

interface ComputeInsightCreationContentProps extends NativeContainerProps {
    initialValue?: Partial<CreateComputeInsightFormFields>

    onChange: (event: FormChangeEvent<CreateComputeInsightFormFields>) => void
    onSubmit: (values: CreateComputeInsightFormFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

export const ComputeInsightCreationContent: FunctionComponent<ComputeInsightCreationContentProps> = props => {
    const { initialValue, onChange, onSubmit, onCancel, ...attributes } = props

    return (
        <CreationUiLayout {...attributes}>
            <CreationUIForm>Hello World</CreationUIForm>

            <CreationUIPreview>This is live preview</CreationUIPreview>
        </CreationUiLayout>
    )
}
