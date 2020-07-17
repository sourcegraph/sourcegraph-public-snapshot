import React, { useEffect, useState, useCallback } from 'react'
import classNames from 'classnames'
import { initial } from 'lodash'

interface Props {
    disabled: boolean
    disabledTooltip: string
    message: React.ReactFragment

    buttonText: string
    onButtonClick: (closeChangesets: boolean) => void
    buttonClassName?: string

    /** Used for visual testing. */
    initiallyOpen?: boolean
}

/**
 * A prompt for closing or deleting a campaign, with an option to also close all associated
 * changesets.
 */
export const CloseDeleteCampaignPrompt: React.FunctionComponent<Props> = ({
    disabled,
    disabledTooltip,
    message,
    buttonText,
    onButtonClick,
    buttonClassName = '',
    initiallyOpen = false,
}) => {
    const detailsMenuReference = React.createRef<HTMLDetailsElement>()
    useEffect(() => {
        if (initiallyOpen && detailsMenuReference.current) {
            detailsMenuReference.current.open = true
        }
    }, [initiallyOpen, detailsMenuReference])
    // Global click event listener, used for detecting interaction with other elements. Closes the menu then.
    useEffect(() => {
        const listener = (event: MouseEvent): void => {
            if (!detailsMenuReference.current || !event.target || !(event.target instanceof HTMLElement)) {
                return
            }
            // Only close if nothing within the details menu was clicked
            if (event.target !== detailsMenuReference.current && !detailsMenuReference.current.contains(event.target)) {
                detailsMenuReference.current.open = false
            }
        }
        document.addEventListener('click', listener)
        return () => document.removeEventListener('click', listener)
    }, [detailsMenuReference])
    const [closeChangesets, setCloseChangesets] = useState<boolean>(false)
    const onClick = useCallback(() => onButtonClick(closeChangesets), [onButtonClick, closeChangesets])
    return (
        <details className="campaign-prompt__details" ref={detailsMenuReference}>
            <summary>
                <span
                    className={classNames('btn dropdown-toggle', buttonClassName, disabled && 'disabled')}
                    onClick={event => disabled && event.preventDefault()}
                    data-tooltip={disabled ? disabledTooltip : undefined}
                >
                    {buttonText}
                </span>
            </summary>
            <div className="position-absolute campaign-prompt__details-menu">
                <div className="card mt-1">
                    <div className="card-body">
                        {message}
                        <div className="form-group">
                            <label>
                                <input
                                    type="checkbox"
                                    checked={closeChangesets}
                                    onChange={event => setCloseChangesets(event.target.checked)}
                                />{' '}
                                Close open changesets on code hosts
                            </label>
                        </div>
                        <button
                            type="button"
                            disabled={disabled}
                            className={`btn mr-1 ${buttonClassName}`}
                            onClick={onClick}
                        >
                            {buttonText}
                        </button>
                    </div>
                </div>
            </div>
        </details>
    )
}
