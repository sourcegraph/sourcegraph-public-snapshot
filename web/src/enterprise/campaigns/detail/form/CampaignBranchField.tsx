import React from 'react'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'

interface Props {
    value: string | undefined
    onChange: (newValue: string) => void

    disabled?: boolean
}

/**
 * A text field for a campaign's branch.
 */
export const CampaignBranchField: React.FunctionComponent<Props> = ({ value, onChange, disabled }) => (
    <div className="form-group">
        <label htmlFor="campaignBranch">
            Branch name{' '}
            <small>
                <InformationOutlineIcon
                    className="icon-inline"
                    data-tooltip={
                        'If a branch with the given name already exists, a fallback name will be created by appending a count. Example: "my-branch-name" becomes "my-branch-name-1".'
                    }
                />
            </small>
        </label>
        <input
            id="campaignBranch"
            type="text"
            className="form-control"
            value={value}
            onChange={event => onChange(event.target.value)}
            required={true}
            disabled={disabled}
        />
    </div>
)
