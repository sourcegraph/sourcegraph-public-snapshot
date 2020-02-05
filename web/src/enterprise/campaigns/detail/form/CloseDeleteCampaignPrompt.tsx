import React, { useEffect, useState, useCallback } from 'react'
import { pluralize } from '../../../../../../shared/src/util/strings'

interface Props {
    summary: React.ReactFragment
    message: React.ReactFragment

    changesetsCount: number

    buttonText: string
    onButtonClick: (closeChangesets: boolean) => void
    buttonClassName?: string
    buttonDisabled?: boolean
}

/**
 * A prompt for closing or deleting a campaign, with an option to also close all associated
 * changesets.
 */
export const CloseDeleteCampaignPrompt: React.FunctionComponent<Props> = ({
    summary,
    message,
    changesetsCount,
    buttonText,
    onButtonClick,
    buttonClassName = '',
    buttonDisabled,
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
    const [closeChangesets, setCloseChangesets] = useState<boolean>(false)
    const onClick = useCallback(() => onButtonClick(closeChangesets), [onButtonClick, closeChangesets])
    return (
        <>
            <details className="campaign-prompt__details" ref={detailsMenuRef}>
                <summary>{summary}</summary>
                <div className="position-absolute campaign-prompt__details-menu">
                    <div className="card mt-1">
                        <div className="card-body">
                            {message}
                            <div className="form-group">
                                <label>
                                    <input
                                        type="checkbox"
                                        checked={closeChangesets}
                                        onChange={e => setCloseChangesets(e.target.checked)}
                                    />{' '}
                                    Close all {changesetsCount} {pluralize('changeset', changesetsCount)} on code hosts
                                </label>
                            </div>
                            <button
                                type="button"
                                disabled={buttonDisabled}
                                className={`btn mr-1 ${buttonClassName}`}
                                onClick={onClick}
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
