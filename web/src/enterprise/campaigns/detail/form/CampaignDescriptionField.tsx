import React from 'react'

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
    onChange,
    className = '',
    disabled,
}) => (
    <textarea
        className={`form-control ${className}`}
        value={value}
        onChange={event => onChange(event.target.value)}
        placeholder="Description (purpose of campaign, instructions for reviewers, links to relevant internal documentation, etc.)"
        rows={8}
        disabled={disabled}
    />
)
