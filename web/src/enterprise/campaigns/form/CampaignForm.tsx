import React, { useCallback } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { CampaignFormCommonFields } from './CampaignFormCommonFields'

export interface CampaignFormData extends GQL.ICreateCampaignInput {
    isValid: boolean
}

export interface CampaignFormControl {
    value: CampaignFormData
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
> = ({ value, onChange, onSubmit: parentOnSubmit, disabled, isLoading, className = '', children }) => {
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
                        onChange={onChange}
                        disabled={disabled}
                        isLoading={isLoading}
                        className="mt-4"
                    />
                ),
            })}
        </Form>
    )
}
