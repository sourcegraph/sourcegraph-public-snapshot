import React, { useCallback } from 'react'
import { Action } from '../../../../../shared/src/api/types/action'

interface Props {
    /** The action. */
    action: Action

    /** Called when the button is clicked. */
    onClick: (action: Action) => void

    disabled?: boolean
    className?: string
}

/**
 * A button for a command action that can be invoked.
 *
 * TODO!(sqs): use <ActionItem /> (same as for action contributions)?
 */
export const CommandActionButton: React.FunctionComponent<Props> = ({ action, onClick, disabled, className = '' }) => {
    const onButtonClick = useCallback(() => onClick(action), [action, onClick])

    const title = action.title || (action.command ? action.command.title : undefined)

    return (
        <button
            type="button"
            onClick={onButtonClick}
            disabled={disabled}
            className={className}
            data-tooltip={action.command ? action.command.tooltip : undefined}
        >
            {title || '(Untitled action)'}
        </button>
    )
}
