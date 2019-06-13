import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback } from 'react'
import { Action } from '../../../../../shared/src/api/types/action'

interface Props {
    /** The action. */
    action: Action

    /**
     * Whether the action is active.
     */
    value: boolean

    /**
     * Called when the active/inactive value of this action is changed by user interaction.
     */
    onChange: (value: boolean, action: Action) => void

    className?: string
    buttonClassName?: string
    activeButtonClassName?: string
    inactiveButtonClassName?: string
    disabled?: boolean
}

/**
 * A radio button for a single action that can be set as active (instead of just being invoked).
 */
export const ActionRadioButton: React.FunctionComponent<Props> = ({
    action,
    value,
    onChange,
    className,
    buttonClassName = 'btn btn-link text-decoration-none',
    activeButtonClassName,
    inactiveButtonClassName,
    disabled,
}) => {
    const onInputChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange(e.currentTarget.checked, action)
        },
        [action, onChange]
    )
    const onClear = useCallback(() => onChange(false, action), [action, onChange])
    return (
        <div className={`action-radio-button ${className}`}>
            <label
                className={`${buttonClassName} ${value ? activeButtonClassName : inactiveButtonClassName}`}
                style={{ cursor: 'pointer' }}
            >
                <input type="radio" className="mr-2" checked={value} onChange={onInputChange} disabled={disabled} />
                {action.title}
            </label>
            {value && (
                <button
                    className={`${buttonClassName} ${inactiveButtonClassName} btn-sm text-muted`}
                    onClick={onClear}
                    disabled={disabled}
                >
                    <CloseIcon className="icon-inline" /> Clear
                </button>
            )}
        </div>
    )
}
