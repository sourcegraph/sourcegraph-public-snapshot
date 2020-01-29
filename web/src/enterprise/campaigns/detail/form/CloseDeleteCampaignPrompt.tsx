import React, { useEffect } from 'react'
import { pluralize } from '../../../../../../shared/src/util/strings'

interface Props {
    summary: React.ReactFragment
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
    summary,
    message,
    changesetsCount,
    closeChangesets,
    onCloseChangesetsToggle,
    buttonText,
    onButtonClick,
    buttonClassName = '',
    buttonDisabled,
    className,
}) => {
    const detailsMenuRef = React.createRef<HTMLDetailsElement>()
    // Global click event listener, used for detecting interaction with other elements. Closes the menu then.
    useEffect(() => {
        const listener = (event: MouseEvent): void => {
            if (!detailsMenuRef.current || !event.target || !(event.target instanceof HTMLElement)) {
                return
            }
            // Only close if nothing within the details menu was clicked
            if (event.target !== detailsMenuRef.current && !detailsMenuRef.current.contains(event.target)) {
                detailsMenuRef.current.open = false
            }
        }
        document.addEventListener('click', listener)
        return () => document.removeEventListener('click', listener)
    }, [detailsMenuRef])
    return (
        <>
            <details className="campaign-prompt__details" ref={detailsMenuRef}>
                <summary>{summary}</summary>
                <div className={className}>
                    <div className="card mt-1">
                        <div className="card-body">
                            {message}
                            <div className="form-group">
                                <input
                                    id={`checkbox-campaign-prompt-${buttonText}`}
                                    type="checkbox"
                                    checked={closeChangesets}
                                    onChange={e => onCloseChangesetsToggle(e.target.checked)}
                                />{' '}
                                <label htmlFor={`checkbox-campaign-prompt-${buttonText}`}>
                                    Close all {changesetsCount} {pluralize('changeset', changesetsCount)} on code hosts
                                </label>
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
            </details>
        </>
    )
}
