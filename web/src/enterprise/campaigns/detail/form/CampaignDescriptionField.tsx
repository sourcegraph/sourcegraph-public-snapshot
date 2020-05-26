import React, { useCallback } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'

interface Props {
    value: string | undefined
    onChange: (newValue: string) => void

    disabled?: boolean
}

/**
 * A multi-line text field for a campaign's description.
 */
export const CampaignDescriptionField: React.FunctionComponent<Props> = ({
    value,
    onChange: parentOnChange,
    disabled,
}) => {
    const onChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        event => parentOnChange(event.target.value),
        [parentOnChange]
    )
    return (
        <div className="form-group">
            <label htmlFor="campaignDescription">
                Description{' '}
                <small>
                    <InformationOutlineIcon
                        className="icon-inline"
                        data-tooltip="Purpose of campaign, instructions for reviewers, links to relevant internal documentation, etc."
                    />
                </small>
            </label>
            <TextareaAutosize
                type="text"
                className="form-control"
                value={value}
                onChange={onChange}
                minRows={3}
                disabled={disabled}
                id="campaignDescription"
            />
            <p className="ml-1">
                <small>
                    <a rel="noopener noreferrer" target="_blank" href="/help/user/markdown" tabIndex={-1}>
                        Markdown supported
                    </a>
                </small>
            </p>
        </div>
    )
}
