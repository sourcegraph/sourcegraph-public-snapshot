import React, { useCallback } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import { CampaignFormControl } from './CampaignForm'

interface Props extends CampaignFormControl {
    disabled?: boolean
    className?: string
}

/**
 * The common form fields for campaigns.
 */
export const CampaignFormCommonFields: React.FunctionComponent<Props> = ({
    value,
    onChange,
    disabled,
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
                <label htmlFor="campaign-form-common-fields__name">Name</label>
                <input
                    type="text"
                    id="campaign-form-common-fields__name"
                    className="form-control"
                    required={true}
                    minLength={1}
                    placeholder="Campaign name"
                    value={name}
                    onChange={onNameChange}
                    autoFocus={true}
                    disabled={disabled}
                />
            </div>
            <div className="form-group">
                <label htmlFor="campaign-form-common-fields__body">Body</label>
                <TextareaAutosize
                    type="text"
                    id="campaign-form-common-fields__body"
                    className="form-control"
                    placeholder="Body"
                    value={value.body || ''}
                    minRows={3}
                    onChange={onBodyChange}
                    disabled={disabled}
                />
            </div>
        </div>
    )
}
