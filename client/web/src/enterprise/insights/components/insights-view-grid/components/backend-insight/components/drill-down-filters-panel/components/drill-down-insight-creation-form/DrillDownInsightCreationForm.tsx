import classnames from 'classnames'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../../../../../../../../components/alerts'
import { LoaderButton } from '../../../../../../../../../../components/LoaderButton'
import { FormInput } from '../../../../../../../form/form-input/FormInput'
import { useAsyncInsightTitleValidator } from '../../../../../../../form/hooks/use-async-insight-title-validator'
import { useField } from '../../../../../../../form/hooks/useField'
import { FORM_ERROR, SubmissionResult, useForm } from '../../../../../../../form/hooks/useForm'
import { createRequiredValidator } from '../../../../../../../form/validators'

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

export const DrillDownInsightCreationForm: React.FunctionComponent<DrillDownInsightCreationFormProps> = props => {
    const { className, onCreateInsight, onCancel } = props

    const { formAPI, ref, handleSubmit } = useForm({
        initialValues: DEFAULT_FORM_VALUES,
        onSubmit: onCreateInsight,
    })

    const titleDuplicationValidator = useAsyncInsightTitleValidator({
        initialTitle: '',
    })

    const insightName = useField({
        name: 'insightName',
        formApi: formAPI,
        validators: { sync: insightRequiredValidator, async: titleDuplicationValidator },
    })

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} onSubmit={handleSubmit} noValidate={true} className={classnames(className, 'p-3')}>
            <h3 className="mb-3">Save as new view</h3>

            <FormInput
                title="Name"
                autoFocus={true}
                required={true}
                description="Shown as the title for your insight"
                placeholder="Example: Migration to React function components"
                valid={insightName.meta.touched && insightName.meta.validState === 'VALID'}
                error={insightName.meta.touched && insightName.meta.error}
                {...insightName.input}
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
                    className="btn btn-primary"
                />
            </footer>
        </form>
    )
}
