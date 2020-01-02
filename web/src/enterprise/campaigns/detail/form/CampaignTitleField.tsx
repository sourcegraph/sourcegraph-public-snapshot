import React from 'react'

interface Props {
    value: string | undefined
    onChange: (newValue: string) => void

    className?: string
    disabled?: boolean
}

/**
 * A text field for a campaign's title.
 */
export const CampaignTitleField: React.FunctionComponent<Props> = ({ value, onChange, className = '', disabled }) => (
    <input
        className={`form-control ${className}`}
        value={value}
        onChange={event => onChange(event.target.value)}
        placeholder="Campaign title"
        disabled={disabled}
        autoFocus={true}
        required={true}
    />
)
