import classnames from 'classnames'
import React, { useMemo } from 'react'

import { Settings } from '@sourcegraph/shared/src/settings/settings'
import { Button } from '@sourcegraph/wildcard'

import { ErrorAlert } from '../../../../../../../../../../components/alerts'
import { LoaderButton } from '../../../../../../../../../../components/LoaderButton'
import { InsightTypePrefix } from '../../../../../../../../core/types'
import { FormInput } from '../../../../../../../form/form-input/FormInput'
import { useField, Validator } from '../../../../../../../form/hooks/useField'
import { FORM_ERROR, SubmissionResult, useForm } from '../../../../../../../form/hooks/useForm'
import {
    useInsightTitleDuplicationCheck,
    useTitleValidatorProps,
} from '../../../../../../../form/hooks/useInsightTitleValidator'
import { composeValidators, createRequiredValidator } from '../../../../../../../form/validators'

export interface DrillDownInsightCreationFormValues {
    insightName: string
}

const DEFAULT_FORM_VALUES: DrillDownInsightCreationFormValues = {
    insightName: '',
}

function useInsightNameValidator(props: useTitleValidatorProps): Validator<string> {
    const hasInsightTitleDuplication = useInsightTitleDuplicationCheck(props)

    return useMemo(
        () =>
            composeValidators<string>(
                createRequiredValidator('Insight name is a required field.'),
                hasInsightTitleDuplication
            ),
        [hasInsightTitleDuplication]
    )
}

interface DrillDownInsightCreationFormProps {
    className?: string
    settings: Settings
    onCreateInsight: (values: DrillDownInsightCreationFormValues) => SubmissionResult
    onCancel: () => void
}

export const DrillDownInsightCreationForm: React.FunctionComponent<DrillDownInsightCreationFormProps> = props => {
    const { settings, className, onCreateInsight, onCancel } = props

    const { formAPI, ref, handleSubmit } = useForm({
        initialValues: DEFAULT_FORM_VALUES,
        onSubmit: onCreateInsight,
    })

    const nameValidator = useInsightNameValidator({
        insightType: InsightTypePrefix.search,
        settings,
    })

    const insightName = useField({
        name: 'insightName',
        formApi: formAPI,
        validators: { sync: nameValidator },
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
