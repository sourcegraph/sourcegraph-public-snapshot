import React, { useCallback } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import { CampaignFormControl } from './CampaignForm'

interface Props extends CampaignFormControl {
    autoFocus?: boolean
    className?: string
}

/**
 * The common form fields for campaigns.
 */
export const CampaignFormCommonFields: React.FunctionComponent<Props> = ({
    value,
    onChange,
    disabled,
    autoFocus,
    className = '',
}) => {
    const onNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => onChange({ ...value, name: e.currentTarget.value }),
        [onChange, value]
    )

    const onBodyChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => onChange({ ...value, body: e.currentTarget.value }),
        [onChange, value]
    )

    return (
        <div className={className}>
            <div className="form-group">
                <label htmlFor="campaign-form-common-fields__name">Campaign name</label>
                <input
                    type="text"
                    id="campaign-form-common-fields__name"
                    className="form-control"
                    required={true}
                    minLength={1}
                    placeholder="Campaign name"
                    value={value.name}
                    onChange={onNameChange}
                    autoFocus={autoFocus}
                    disabled={disabled}
                />
            </div>
            <div className="form-group">
                <label htmlFor="campaign-form-common-fields__body">Campaign description</label>
                <TextareaAutosize
                    type="text"
                    id="campaign-form-common-fields__body"
                    className="form-control"
                    placeholder="Describe the purpose of this campaign, link to relevant internal documentation, etc."
                    value={value.body || ''}
                    minRows={3}
                    onChange={onBodyChange}
                    disabled={disabled}
                />
            </div>
        </div>
    )
}
