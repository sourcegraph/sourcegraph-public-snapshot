import React, { useCallback } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { useLocalStorage } from '../../../util/useLocalStorage'
import { CampaignFormCommonFields } from './CampaignFormCommonFields'

export interface CampaignFormData
    extends Pick<GQL.ICreateCampaignInput, Exclude<keyof GQL.ICreateCampaignInput, 'extensionData'>> {
    isValid: boolean
    draft: boolean
    startDate?: string // TODO!(sqs): implement
    dueDate?: string // TODO!(sqs): implement
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

    const [isCreateCampaignInputVisible, setIsCreateCampaignInputVisible] = useLocalStorage(
        'CampaignForm.isCreateCampaignInputVisible',
        false
    )
    const toggleIsCreateCampaignInputVisible = useCallback(
        () => setIsCreateCampaignInputVisible(!isCreateCampaignInputVisible),
        [isCreateCampaignInputVisible, setIsCreateCampaignInputVisible]
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
            <div className="form-group mt-4">
                <button type="button" className="btn btn-sm btn-link ml-2" onClick={toggleIsCreateCampaignInputVisible}>
                    {isCreateCampaignInputVisible ? 'Hide' : 'Show'} JSON
                </button>
            </div>
            {isCreateCampaignInputVisible && (
                <pre className="small mt-4 border p-2 overflow-auto">
                    <code>{JSON.stringify(value, null, 2)}</code>
                </pre>
            )}
        </Form>
    )
}
