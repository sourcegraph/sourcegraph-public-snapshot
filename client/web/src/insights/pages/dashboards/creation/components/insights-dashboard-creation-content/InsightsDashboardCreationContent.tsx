import React from 'react'

import { ErrorAlert } from '../../../../../../components/alerts'
import { LoaderButton } from '../../../../../../components/LoaderButton'
import { Settings } from '../../../../../../schema/settings.schema'
import { FormGroup } from '../../../../../components/form/form-group/FormGroup'
import { FormInput } from '../../../../../components/form/form-input/FormInput'
import { FormRadioInput } from '../../../../../components/form/form-radio-input/FormRadioInput'
import { useField } from '../../../../../components/form/hooks/useField'
import { FORM_ERROR, SubmissionErrors, useForm } from '../../../../../components/form/hooks/useForm'
import { Organization } from '../../../../../components/visibility-picker/VisibilityPicker'

import { useDashboardNameValidator } from './hooks/useDashboardNameValidator'

const DASHBOARD_INITIAL_VALUES: DashboardCreationFields = {
    name: '',
    visibility: 'personal',
}

export interface DashboardCreationFields {
    name: string
    visibility: string
}

export interface InsightsDashboardCreationContentProps {
    /**
     * Initial values for the dashboard creation form.
     */
    initialValues?: DashboardCreationFields

    /**
     * Organizations list used in the creation form for dashboard visibility setting.
     */
    organizations: Organization[]

    settings: Settings

    onSubmit: (values: DashboardCreationFields) => SubmissionErrors | Promise<SubmissionErrors> | void
    onCancel: () => void
}

/**
 * Renders creation UI form content (fields, submit and cancel buttons).
 */
export const InsightsDashboardCreationContent: React.FunctionComponent<InsightsDashboardCreationContentProps> = props => {
    const { initialValues = DASHBOARD_INITIAL_VALUES, organizations, settings, onCancel, onSubmit } = props

    const { ref, handleSubmit, formAPI } = useForm<DashboardCreationFields>({
        initialValues,
        onSubmit,
    })

    const nameValidator = useDashboardNameValidator({ settings })
    const name = useField('name', formAPI, { sync: nameValidator })
    const visibility = useField('visibility', formAPI)

    return (
        // eslint-disable-next-line react/forbid-elements
        <form noValidate={true} ref={ref} onSubmit={handleSubmit}>
            <FormInput
                required={true}
                autoFocus={true}
                title="Name"
                placeholder="Example: My personal code insight dashboard"
                description="Shown as the title for your dashboard"
                valid={name.meta.touched && name.meta.validState === 'VALID'}
                error={name.meta.touched && name.meta.error}
                {...name.input}
            />

            <FormGroup name="visibility" title="Visibility" className="mb-0 mt-4">
                <FormRadioInput
                    name="visibility"
                    value="personal"
                    title="Private"
                    description="visible only to you"
                    checked={visibility.input.value === 'personal'}
                    className="mr-3"
                    onChange={visibility.input.onChange}
                />

                <hr className="mt-2 mb-3" />

                <small className="d-block text-muted mb-3">
                    Shared - visible to everyone is the chosen Organisation
                </small>

                {organizations.map(org => (
                    <FormRadioInput
                        key={org.id}
                        name="visibility"
                        value={org.id}
                        title={org.displayName ?? org.name}
                        checked={visibility.input.value === org.id}
                        onChange={visibility.input.onChange}
                        className="mr-3"
                    />
                ))}

                {organizations.length === 0 && (
                    <FormRadioInput
                        name="visibility"
                        value="organization"
                        disabled={true}
                        title="Organization"
                        description="all users in your organization"
                        labelTooltipPosition="right"
                        className="d-inline-block mr-3"
                        labelTooltipText="Create or join an organization to share the dashboard with others!"
                    />
                )}
            </FormGroup>

            {formAPI.submitErrors?.[FORM_ERROR] && (
                <ErrorAlert error={formAPI.submitErrors[FORM_ERROR]} className="mt-2 mb-2" />
            )}

            <div className="d-flex flex-wrap justify-content-end mt-3">
                <button type="button" className="btn btn-outline-secondary mb-2" onClick={onCancel}>
                    Cancel
                </button>

                <LoaderButton
                    alwaysShowLabel={true}
                    data-testid="insight-save-button"
                    loading={formAPI.submitting}
                    label={formAPI.submitting ? 'Submitting' : 'Create dashboard'}
                    type="submit"
                    disabled={formAPI.submitting && formAPI.touched}
                    className="btn btn-primary ml-2 mb-2"
                />
            </div>
        </form>
    )
}
