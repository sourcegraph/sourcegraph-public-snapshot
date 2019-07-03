import React, { useCallback } from 'react'
import { CodeAction } from 'sourcegraph'

interface Props {
    /** The action. */
    action: CodeAction

    /** Called when the button is clicked. */
    onClick: (action: CodeAction) => void

    className?: string
}

/**
 * A button for a single action that can be invoked.
 */
export const ActionButton: React.FunctionComponent<Props> = ({ action, onClick, className = '' }) => {
    const onButtonClick = useCallback(() => onClick(action), [action, onClick])
    return (
        <button type="button" className={className} onClick={onButtonClick}>
            {action.title}
        </button>
    )
}
