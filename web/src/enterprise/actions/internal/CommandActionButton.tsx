import React, { useCallback } from 'react'
import { Action, ActionType } from '../../../../../shared/src/api/types/action'

interface Props {
    /** The action. */
    action: ActionType['command']

    /** Called when the button is clicked. */
    onClick: (action: Action) => void

    disabled?: boolean
    className?: string
}

/**
 * A button for a command action that can be invoked.
 */
export const CommandActionButton: React.FunctionComponent<Props> = ({ action, onClick, disabled, className = '' }) => {
    const onButtonClick = useCallback(() => onClick(action), [action, onClick])
    return (
        <button
            type="button"
            onClick={onButtonClick}
            disabled={disabled}
            className={className}
            data-tooltip={action.command.tooltip}
        >
            {action.command.title}
        </button>
    )
}
