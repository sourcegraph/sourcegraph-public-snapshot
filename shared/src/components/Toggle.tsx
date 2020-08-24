import * as React from 'react'

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

/** A toggle switch input component. */
export const Toggle: React.FunctionComponent<Props> = ({
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
            className={`toggle ${disabled ? 'toggle__disabled' : ''} ${className || ''}`}
            id={id}
            title={title}
            value={value ? 1 : 0}
            onClick={onClick}
            tabIndex={tabIndex}
        >
            <span
                className={`toggle__bar ${value ? 'toggle__bar--active' : ''} ${
                    disabled ? 'toggle__bar--disabled' : ''
                }`}
            />
            <span
                className={`toggle__knob ${value ? 'toggle__knob--active' : ''} ${
                    disabled ? (value ? 'toggle__knob--disabled__active' : 'toggle__knob--disabled') : ''
                }`}
            />
        </button>
    )
}
