import React from 'react';

import { Button } from '@sourcegraph/wildcard';

import { ErrorAlert } from '../../../../../../../../../components/alerts';
import { LoaderButton } from '../../../../../../../../../components/LoaderButton';
import { InsightTypePrefix } from '../../../../../../../../core/types';
import { FormInput } from '../../../../../../../form/form-input/FormInput'
import { useField } from '../../../../../../../form/hooks/useField';
import { FORM_ERROR, SubmissionResult, useForm } from '../../../../../../../form/hooks/useForm';
import { useInsightTitleValidator } from '../../../../../../../form/hooks/useInsightTitleValidator';

interface DrillDownInsightCreationFormValues {
    insightName: string
}

const DEFAULT_FORM_VALUES: DrillDownInsightCreationFormValues = {
    insightName: ''
}

interface DrillDownInsightCreationFormProps {
    onCreateInsight: (values: DrillDownInsightCreationFormValues) => SubmissionResult
    onReset: () => void
}

export const DrillDownInsightCreationForm: React.FunctionComponent<DrillDownInsightCreationFormProps> = props => {
    const { onCreateInsight, onReset } = props;

    const { formAPI, ref, handleSubmit, } = useForm({
        initialValues: DEFAULT_FORM_VALUES,
        onSubmit: onCreateInsight
    })

    const nameValidator = useInsightTitleValidator({ insightType: InsightTypePrefix.search, settings: null })

    const insightName = useField({
        name: 'insightName',
        formApi: formAPI,
        validators: { sync: nameValidator }
    })

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} onSubmit={handleSubmit} noValidate={true}>
            <h4>Save as new view</h4>

            <FormInput
                title="Name"
                required={true}
                description="Shown as the title for your insight"
                placeholder="Example: Migration to React function components"
                valid={insightName.meta.touched && insightName.meta.validState === 'VALID'}
                error={insightName.meta.touched && insightName.meta.error}
                {...insightName.input}
            />

            <footer className='d-flex flex-wrap align-items-baseline'>
                {formAPI.submitErrors?.[FORM_ERROR] &&
                    <ErrorAlert className='' error={formAPI.submitErrors[FORM_ERROR]} />}

                <Button type="reset" variant='secondary' onClick={onReset}>
                    Cancel
                </Button>

                <LoaderButton
                    type="submit"
                    alwaysShowLabel={true}
                    loading={formAPI.submitting}
                    label={formAPI.submitting ? 'Saving' : 'Save'}
                    disabled={formAPI.submitting}
                    data-testid="insight-save-button"
                    className="btn btn-primary mr-2 mb-2"
                />
            </footer>
        </form>
    )
}
