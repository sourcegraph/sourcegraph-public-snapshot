import React from 'react'

interface Props {
    value: string | undefined
    onChange: (newValue: string) => void

    disabled?: boolean
}

/**
 * A text field for a campaign's title.
 */
export const CampaignTitleField: React.FunctionComponent<Props> = ({ value, onChange, disabled }) => (
    <div className="form-group">
        <label htmlFor="campaignTitle">Title</label>
        <input
            className="form-control test-campaign-title"
            value={value}
            onChange={event => onChange(event.target.value)}
            disabled={disabled}
            autoFocus={true}
            required={true}
            id="campaignTitle"
        />
    </div>
)
