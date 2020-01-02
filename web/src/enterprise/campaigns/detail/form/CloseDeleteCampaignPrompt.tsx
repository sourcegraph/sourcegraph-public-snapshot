import React from 'react'
import { pluralize } from '../../../../../../shared/src/util/strings'

interface Props {
    message: React.ReactFragment

    changesetsCount: number
    closeChangesets: boolean
    onCloseChangesetsToggle: (newValue: boolean) => void

    buttonText: string
    onButtonClick: () => void
    buttonClassName?: string
    buttonDisabled?: boolean

    className?: string
}

/**
 * A prompt for closing or deleting a campaign, with an option to also close all associated
 * changesets.
 */
export const CloseDeleteCampaignPrompt: React.FunctionComponent<Props> = ({
    message,
    changesetsCount,
    closeChangesets,
    onCloseChangesetsToggle,
    buttonText,
    onButtonClick,
    buttonClassName = '',
    buttonDisabled,
    className,
}) => (
    <div className={className}>
        <div className="card mt-1">
            <div className="card-body">
                {message}
                <div className="form-group">
                    <input
                        type="checkbox"
                        checked={closeChangesets}
                        onChange={e => onCloseChangesetsToggle(e.target.checked)}
                    />{' '}
                    Close all {changesetsCount} {pluralize('changeset', changesetsCount)} on code hosts
                </div>
                <button
                    type="button"
                    disabled={buttonDisabled}
                    className={`btn mr-1 ${buttonClassName}`}
                    onClick={onButtonClick}
                >
                    {buttonText}
                </button>
            </div>
        </div>
    </div>
)
