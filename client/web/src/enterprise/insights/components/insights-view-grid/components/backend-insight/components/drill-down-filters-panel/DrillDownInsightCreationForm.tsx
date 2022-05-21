import { FunctionComponent } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, FormInput, Typography } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../../../../../components/LoaderButton'
import { getDefaultInputProps } from '../../../../../form/getDefaultInputProps'
import { useAsyncInsightTitleValidator } from '../../../../../form/hooks/use-async-insight-title-validator'
import { useField } from '../../../../../form/hooks/useField'
import { FORM_ERROR, SubmissionResult, useForm } from '../../../../../form/hooks/useForm'
import { createRequiredValidator } from '../../../../../form/validators'

export interface DrillDownInsightCreationFormValues {
    insightName: string
}

const insightRequiredValidator = createRequiredValidator('Insight name is a required field.')

const DEFAULT_FORM_VALUES: DrillDownInsightCreationFormValues = {
    insightName: '',
}

interface DrillDownInsightCreationFormProps {
    className?: string
    onCreateInsight: (values: DrillDownInsightCreationFormValues) => SubmissionResult
    onCancel: () => void
}

export const DrillDownInsightCreationForm: FunctionComponent<DrillDownInsightCreationFormProps> = props => {
    const { className, onCreateInsight, onCancel } = props

    const { formAPI, ref, handleSubmit } = useForm({
        initialValues: DEFAULT_FORM_VALUES,
        onSubmit: onCreateInsight,
    })

    const titleDuplicationValidator = useAsyncInsightTitleValidator({
        initialTitle: '',
        mode: 'creation',
    })

    const insightName = useField({
        name: 'insightName',
        formApi: formAPI,
        validators: { sync: insightRequiredValidator, async: titleDuplicationValidator },
    })

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} onSubmit={handleSubmit} noValidate={true} className={className}>
            <Typography.H3 className="mb-3">Save as new view</Typography.H3>

            <FormInput
                label="Name"
                autoFocus={true}
                required={true}
                message="Shown as the title for your insight"
                placeholder="Example: Migration to React function components"
                {...getDefaultInputProps(insightName)}
            />

            <footer className="mt-4 d-flex flex-wrap align-items-center">
                {formAPI.submitErrors?.[FORM_ERROR] && (
                    <ErrorAlert className="w-100 mb-3" error={formAPI.submitErrors[FORM_ERROR]} />
                )}

                <Button type="reset" variant="secondary" className="ml-auto mr-2" onClick={onCancel}>
                    Cancel
                </Button>

                <LoaderButton
                    type="submit"
                    alwaysShowLabel={true}
                    loading={formAPI.submitting}
                    label={formAPI.submitting ? 'Saving' : 'Save'}
                    disabled={formAPI.submitting}
                    data-testid="insight-save-button"
                    variant="primary"
                />
            </footer>
        </form>
    )
}
