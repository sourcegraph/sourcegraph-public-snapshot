import React, { ReactNode, useContext } from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Input } from '@sourcegraph/wildcard'

import { FormGroup } from '../../../../components/form/form-group/FormGroup'
import { FormRadioInput } from '../../../../components/form/form-radio-input/FormRadioInput'
import { getDefaultInputProps } from '../../../../components/form/getDefaultInputProps'
import { useField } from '../../../../components/form/hooks/useField'
import { FORM_ERROR, FormAPI, SubmissionErrors, useForm } from '../../../../components/form/hooks/useForm'
import { createRequiredValidator } from '../../../../components/form/validators'
import { LimitedAccessLabel } from '../../../../components/limited-access-label/LimitedAccessLabel'
import {
    CodeInsightsBackendContext,
    InsightsDashboardOwner,
    isGlobalOwner,
    isOrganizationOwner,
    isPersonalOwner,
} from '../../../../core'

import styles from './InsightsDashboardCreationContent.module.scss'

const dashboardTitleRequired = createRequiredValidator('Name is a required field.')

const DASHBOARD_INITIAL_VALUES: DashboardCreationFields = {
    name: '',
    owner: null,
}

export interface DashboardCreationFields {
    name: string
    owner: InsightsDashboardOwner | null
}

export interface InsightsDashboardCreationContentProps {
    initialValues?: DashboardCreationFields
    owners: InsightsDashboardOwner[]
    onSubmit: (values: DashboardCreationFields) => Promise<SubmissionErrors>
    children: (formAPI: FormAPI<DashboardCreationFields>) => ReactNode
}

/**
 * Renders creation UI form content (fields, submit and cancel buttons).
 */
export const InsightsDashboardCreationContent: React.FunctionComponent<
    React.PropsWithChildren<InsightsDashboardCreationContentProps>
> = props => {
    const { initialValues, owners, onSubmit, children } = props

    const { UIFeatures } = useContext(CodeInsightsBackendContext)
    const { licensed } = UIFeatures

    const userOwner = owners.find(isPersonalOwner)
    const personalOwners = owners.filter(isPersonalOwner)
    const organizationOwners = owners.filter(isOrganizationOwner)
    const globalOwners = owners.filter(isGlobalOwner)

    const { ref, handleSubmit, formAPI } = useForm<DashboardCreationFields>({
        initialValues: initialValues ?? {
            ...DASHBOARD_INITIAL_VALUES,
            owner: userOwner ?? null,
        },
        onSubmit,
    })

    const name = useField({
        name: 'name',
        formApi: formAPI,
        validators: { sync: dashboardTitleRequired },
    })

    const visibility = useField({
        name: 'owner',
        formApi: formAPI,
    })

    return (
        // eslint-disable-next-line react/forbid-elements
        <form noValidate={true} ref={ref} onSubmit={handleSubmit}>
            <Input
                required={true}
                autoFocus={true}
                label="Name"
                placeholder="Example: My personal code insights dashboard"
                message="Shown as the title for your dashboard"
                {...getDefaultInputProps(name)}
            />

            <FormGroup name="visibility" title="Visibility" contentClassName="d-flex flex-column" className="mb-0 mt-4">
                {personalOwners.map(owner => (
                    <FormRadioInput
                        key={owner.id}
                        name="visibility"
                        value={owner.id}
                        title="Private"
                        description="visible only to you"
                        checked={visibility.input.value?.id === owner.id}
                        className="mr-3"
                        onChange={() => visibility.input.onChange(owner)}
                    />
                ))}

                <hr className="mt-2 mb-3" />

                <small className="d-block text-muted mb-3">
                    Shared - visible to everyone in the chosen organization
                </small>

                {organizationOwners.map(org => (
                    <FormRadioInput
                        key={org.id}
                        name="visibility"
                        value={org.id}
                        title={org.title}
                        checked={visibility.input.value?.id === org.id}
                        onChange={() => visibility.input.onChange(org)}
                        className="mr-3"
                    />
                ))}

                {organizationOwners.length === 0 && (
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

                {globalOwners.map(owner => (
                    <FormRadioInput
                        key={owner.id}
                        name="visibility"
                        value={owner.id}
                        title={owner.title}
                        description="visible to everyone on your Sourcegraph instance"
                        checked={visibility.input.value?.id === owner.id}
                        className="mr-3 flex-grow-0"
                        onChange={() => visibility.input.onChange(owner)}
                    />
                ))}
            </FormGroup>

            {formAPI.submitErrors?.[FORM_ERROR] && (
                <ErrorAlert error={formAPI.submitErrors[FORM_ERROR]} className="mt-2 mb-2" />
            )}

            {!licensed && (
                <LimitedAccessLabel
                    className={classNames(styles.limitedBanner)}
                    message="Unlock Code Insights to create unlimited custom dashboards"
                />
            )}

            <div className="d-flex flex-wrap justify-content-end mt-3">{children(formAPI)}</div>
        </form>
    )
}
