import React, { useCallback } from 'react'
import TextareaAutosize from 'react-textarea-autosize'

interface Props {
    value: string | undefined
    onChange: (newValue: string) => void

    className?: string
    disabled?: boolean
}

/**
 * A multi-line text field for a campaign's description.
 */
export const CampaignDescriptionField: React.FunctionComponent<Props> = ({
    value,
    onChange: parentOnChange,
    className = '',
    disabled,
}) => {
    const onChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        event => parentOnChange(event.target.value),
        [parentOnChange]
    )
    return (
        <TextareaAutosize
            type="text"
            className={`form-control ${className}`}
            value={value}
            onChange={onChange}
            placeholder="Description (purpose of campaign, instructions for reviewers, links to relevant internal documentation, etc.)"
            minRows={3}
            disabled={disabled}
        />
    )
}
