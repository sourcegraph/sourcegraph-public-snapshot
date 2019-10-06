import React, { useCallback } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { CampaignFormCommonFields } from './CampaignFormCommonFields'

export interface CampaignFormData extends Omit<GQL.IExpCreateCampaignInput, 'extensionData' | 'rules'> {
    draft: boolean
    workflowAsJSONCString: string
    startDate?: string // TODO!(sqs): implement
    dueDate?: string // TODO!(sqs): implement
}

export interface CampaignFormControl {
    value: CampaignFormData
    isValid: boolean
    onChange: (value: Partial<CampaignFormData>) => void
    disabled?: boolean
    isLoading?: boolean
}

export interface CampaignFormProps extends CampaignFormControl {
    /** Called when the form is submitted. */
    onSubmit: () => void

    className?: string
}

/**
 * A form to create or edit a campaign.
 */
export const CampaignForm: React.FunctionComponent<
    CampaignFormProps & { children: ({ form }: { form: React.ReactFragment }) => JSX.Element }
> = ({ value, isValid, onChange, onSubmit: parentOnSubmit, disabled, isLoading, className = '', children }) => {
    const onSubmit = useCallback<React.FormEventHandler>(
        e => {
            e.preventDefault()
            parentOnSubmit()
        },
        [parentOnSubmit]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <style>{'.form-group { max-width: 45rem; }' /* TODO!(sqs): hack */}</style>
            {children({
                form: (
                    <CampaignFormCommonFields
                        value={value}
                        isValid={isValid}
                        onChange={onChange}
                        disabled={disabled}
                        isLoading={isLoading}
                        autoFocus={true}
                        className="mt-4"
                    />
                ),
            })}
        </Form>
    )
}
