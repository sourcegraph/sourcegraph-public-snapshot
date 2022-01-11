import React, { ReactNode, useCallback, useContext } from 'react'

import { asError } from '@sourcegraph/common'

import { ErrorAlert } from '../../../../../../../components/alerts'
import { FormGroup } from '../../../../../components/form/form-group/FormGroup'
import { FormInput } from '../../../../../components/form/form-input/FormInput'
import { FormRadioInput } from '../../../../../components/form/form-radio-input/FormRadioInput'
import { useField } from '../../../../../components/form/hooks/useField'
import { FORM_ERROR, FormAPI, SubmissionErrors, useForm } from '../../../../../components/form/hooks/useForm'
import { AsyncValidator } from '../../../../../components/form/hooks/utils/use-async-validation'
import { createRequiredValidator } from '../../../../../components/form/validators'
import { CodeInsightsBackendContext } from '../../../../../core/backend/code-insights-backend-context'
import {
    isGlobalSubject,
    isOrganizationSubject,
    isUserSubject,
    SupportedInsightSubject,
} from '../../../../../core/types/subjects'

import { getGlobalSubjectTooltipText } from './utils/get-global-subject-tooltip-text'

const dashboardTitleRequired = createRequiredValidator('Name is a required field.')

const DASHBOARD_INITIAL_VALUES: DashboardCreationFields = {
    name: '',
    visibility: 'personal',
}

export interface DashboardCreationFields {
    name: string
    visibility: string
    type?: string
    userId?: string
}

export interface InsightsDashboardCreationContentProps {
    initialValues?: DashboardCreationFields

    /**
     * Organizations list used in the creation form for dashboard visibility setting.
     */
    subjects: SupportedInsightSubject[]

    onSubmit: (values: DashboardCreationFields) => Promise<SubmissionErrors>
    children: (formAPI: FormAPI<DashboardCreationFields>) => ReactNode
}

/**
 * Renders creation UI form content (fields, submit and cancel buttons).
 */
export const InsightsDashboardCreationContent: React.FunctionComponent<InsightsDashboardCreationContentProps> = props => {
    const { initialValues, subjects, onSubmit, children } = props

    const { findDashboardByName } = useContext(CodeInsightsBackendContext)

    // We always have user subject in our settings cascade
    const userSubjectID = subjects.find(isUserSubject)?.id ?? ''
    const organizationSubjects = subjects.filter(isOrganizationSubject)

    // We always have global subject in our settings cascade
    const globalSubject = subjects.find(isGlobalSubject)

    const { ref, handleSubmit, formAPI } = useForm<DashboardCreationFields>({
        initialValues: initialValues ?? { ...DASHBOARD_INITIAL_VALUES, visibility: userSubjectID },
        // Override onSubmit to pass type value
        // to correctly set the grants property for graphql api
        onSubmit: async (): Promise<SubmissionErrors> => {
            let type = 'organization'
            if (visibility.input.value === userSubjectID) {
                type = 'personal'
            }

            if (visibility.input.value === globalSubject?.id) {
                type = 'global'
            }

            return onSubmit({
                name: name.input.value,
                visibility: visibility.input.value,
                type,
            })
        },
    })

    const asyncNameValidator = useCallback<AsyncValidator<string>>(
        async name => {
            // Pass empty value and initial value (for edit page original name is acceptable)
            if (!name || name === '' || name === initialValues?.name) {
                return
            }

            try {
                const possibleDashboard = await findDashboardByName(name).toPromise()

                return possibleDashboard !== null
                    ? 'A dashboard with this name already exists. Please set a different name for the new dashboard.'
                    : undefined
            } catch (error) {
                return asError(error).message || 'Unknown Error'
            }
        },
        [findDashboardByName, initialValues?.name]
    )

    const name = useField({
        name: 'name',
        formApi: formAPI,
        validators: { sync: dashboardTitleRequired, async: asyncNameValidator },
    })

    const visibility = useField({
        name: 'visibility',
        formApi: formAPI,
    })

    return (
        <form noValidate={true} ref={ref} onSubmit={handleSubmit}>
            <FormInput
                required={true}
                autoFocus={true}
                title="Name"
                placeholder="Example: My personal code insights dashboard"
                description="Shown as the title for your dashboard"
                valid={name.meta.touched && name.meta.validState === 'VALID'}
                error={name.meta.touched && name.meta.error}
                {...name.input}
            />

            <FormGroup name="visibility" title="Visibility" contentClassName="d-flex flex-column" className="mb-0 mt-4">
                <FormRadioInput
                    name="visibility"
                    value={userSubjectID}
                    title="Private"
                    description="visible only to you"
                    checked={visibility.input.value === userSubjectID}
                    className="mr-3"
                    onChange={visibility.input.onChange}
                />

                <hr className="mt-2 mb-3" />

                <small className="d-block text-muted mb-3">
                    Shared - visible to everyone in the chosen organization
                </small>

                {organizationSubjects.map(org => (
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

                {organizationSubjects.length === 0 && (
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

                <FormRadioInput
                    name="visibility"
                    value={globalSubject?.id}
                    title="Global"
                    description="visible to everyone on your Sourcegraph instance"
                    checked={visibility.input.value === globalSubject?.id}
                    className="mr-3 flex-grow-0"
                    labelTooltipText={getGlobalSubjectTooltipText(globalSubject)}
                    labelTooltipPosition="bottom"
                    disabled={!globalSubject?.viewerCanAdminister}
                    onChange={visibility.input.onChange}
                />
            </FormGroup>

            {formAPI.submitErrors?.[FORM_ERROR] && (
                <ErrorAlert error={formAPI.submitErrors[FORM_ERROR]} className="mt-2 mb-2" />
            )}

            <div className="d-flex flex-wrap justify-content-end mt-3">{children(formAPI)}</div>
        </form>
    )
}
