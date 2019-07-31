import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useEffect, useState } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'

export interface CampaignFormData extends Pick<GQL.ICampaign, 'name' | 'body'> {}

interface Props {
    initialValue?: CampaignFormData

    /** Called when the form is dismissed with no action taken. */
    onDismiss?: () => void

    /** Called when the form is submitted. */
    onSubmit: (campaign: CampaignFormData) => void

    buttonText: string
    isLoading: boolean

    className?: string
}

/**
 * A form to create or edit a campaign.
 */
export const CampaignForm: React.FunctionComponent<Props> = ({
    initialValue = { name: '', body: null },
    onDismiss,
    onSubmit: onSubmitCampaign,
    buttonText,
    isLoading,
    className = '',
}) => {
    const [name, setName] = useState(initialValue.name)
    const onNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setName(e.currentTarget.value),
        []
    )
    useEffect(() => setName(initialValue.name), [initialValue.name])

    const [body, setBody] = useState(initialValue.body)
    const onBodyChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => setBody(e.currentTarget.value),
        []
    )
    useEffect(() => setBody(initialValue.body), [initialValue.body])

    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            onSubmitCampaign({ name, body })
        },
        [body, name, onSubmitCampaign]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <div className="form-row align-items-end">
                <div className="form-group mb-md-0 col-md-3">
                    <label htmlFor="campaign-form__name">Name</label>
                    <input
                        type="text"
                        id="campaign-form__name"
                        className="form-control"
                        required={true}
                        minLength={1}
                        placeholder="Campaign name"
                        value={name}
                        onChange={onNameChange}
                        autoFocus={true}
                    />
                </div>
                <div className="form-group mb-md-0 col-md-3">
                    <label htmlFor="campaign-form__body">Body</label>
                    <TextareaAutosize
                        type="text"
                        id="campaign-form__body"
                        className="form-control"
                        placeholder="Body"
                        value={body || ''}
                        minRows={3}
                        onChange={onBodyChange}
                    />
                </div>
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
            </div>
        </Form>
    )
}
