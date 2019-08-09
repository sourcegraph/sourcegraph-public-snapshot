import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useEffect, useState } from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Form } from '../../../../components/Form'
import { CampaignFormCommonFields } from './CampaignFormCommonFields'
import { CampaignTemplateChooser } from './CampaignTemplateChooser'
import { CAMPAIGN_TEMPLATES } from './templates'

export interface CampaignFormData
    extends Pick<GQL.ICreateCampaignInput, Exclude<keyof GQL.ICreateCampaignInput, 'namespace'>> {}

export interface CampaignFormControl {
    value: CampaignFormData
    onChange: (value: CampaignFormData) => void
}

interface Props {
    templateID: string | null

    initialValue?: CampaignFormData

    /** Called when the form is dismissed with no action taken. */
    onDismiss?: () => void

    /** Called when the form is submitted. */
    onSubmit: (campaign: CampaignFormData) => void

    buttonText: string
    isLoading: boolean

    className?: string

    /** The URL of the form. */
    match: {
        url: string
    }
}

/**
 * A form to create or edit a campaign.
 */
export const CampaignForm: React.FunctionComponent<Props> = ({
    templateID,
    initialValue = { name: '', body: null },
    onDismiss,
    onSubmit: onSubmitCampaign,
    buttonText,
    isLoading,
    className = '',
    match,
}) => {
    const [value, setValue] = useState<CampaignFormData>(initialValue)
    useEffect(() => setValue(initialValue), [initialValue])

    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            onSubmitCampaign(value)
        },
        [onSubmitCampaign, value]
    )

    const template = (templateID !== null && CAMPAIGN_TEMPLATES.find(({ id }) => id === templateID)) || null
    const formControlProps: CampaignFormControl = {
        value,
        onChange: setValue,
    }

    const urlToFormWithTemplate = useCallback(
        (templateID: string) => `${match.url}?${new URLSearchParams({ template: templateID }).toString()}`,
        [match.url]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            {template === null ? (
                <CampaignTemplateChooser {...formControlProps} urlToFormWithTemplate={urlToFormWithTemplate} />
            ) : (
                template.renderForm({})
            )}
            <CampaignFormCommonFields {...formControlProps} />
            <div className="form-group mb-md-0 col-md-3 text-right">
                {onDismiss && (
                    <button type="reset" className="btn btn-secondary mr-2" onClick={onDismiss}>
                        Cancel
                    </button>
                )}
                <button type="submit" disabled={isLoading} className="btn btn-success">
                    {isLoading ? <LoadingSpinner className="icon-inline" /> : buttonText}
                </button>
            </div>
        </Form>
    )
}
