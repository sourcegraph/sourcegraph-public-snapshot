import * as React from 'react'
import Close from 'mdi-react/CloseIcon'
import Check from 'mdi-react/CheckIcon'

interface Props {
    /** The initial value. */
    value?: boolean

    /** The DOM ID of the element. */
    id?: string

    /**
     * Called when the user changes the input's value.
     */
    onToggle?: (value: boolean) => void

    /**
     * Called when the user clicks the toggle when it is disabled.
     */
    onToggleDisabled?: (value: boolean) => void

    /** The title attribute (tooltip). */
    title?: string

    disabled?: boolean
    tabIndex?: number
    className?: string
}

/** A big toggle switch input component. */
export const ToggleBig: React.FunctionComponent<Props> = ({
    disabled,
    className,
    id,
    title,
    value,
    tabIndex,
    onToggle,
    onToggleDisabled,
}) => {
    function onClick(): void {
        if (!disabled && onToggle) {
            onToggle(!value)
        } else if (disabled && onToggleDisabled) {
            onToggleDisabled(!!value)
        }
    }

    return (
        <button
            type="button"
            className={`toggle-big  ${className || ''}`}
            id={id}
            title={title}
            value={value ? 1 : 0}
            onClick={onClick}
            tabIndex={tabIndex}
        >
            <span className="toggle-big__container">
                <span
                    className={`toggle-big__bar ${value ? 'toggle-big__bar--active' : ''} ${
                        disabled ? 'toggle-big__bar--disabled' : ''
                    }`}
                />
                <span className={`toggle-big__knob ${value ? 'toggle-big__knob--active' : ''}`}>
                    {value ? (
                        <Check size={16} className="toggle-big__icon" />
                    ) : (
                        <Close size={16} className="toggle-big__icon--disabled" />
                    )}
                </span>
                <span className={`toggle-big__text ${!value ? 'toggle-big__text--disabled' : ''}`}>
                    {value ? 'Enabled' : 'Disabled'}
                </span>
            </span>
        </button>
    )
}
