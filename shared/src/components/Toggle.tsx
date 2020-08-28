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
     * Called when the user hovers over the toggle.
     */
    onHover?: (value: boolean) => void

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
    onHover,
}) => {
    function onClick(): void {
        if (!disabled && onToggle) {
            onToggle(!value)
        }
    }

    function onMouseOver(): void {
        if (onHover) {
            onHover(!value)
        }
    }

    return (
        <button
            type="button"
            className={`toggle ${className || ''}`}
            id={id}
            title={title}
            value={value ? 1 : 0}
            onClick={onClick}
            tabIndex={tabIndex}
            onMouseOver={onMouseOver}
        >
            <span
                className={`toggle__bar ${value ? 'toggle__bar--active' : ''} ${
                    disabled ? 'toggle__bar--disabled' : ''
                }`}
            />
            <span
                className={`toggle__knob ${value ? 'toggle__knob--active' : ''} ${
                    disabled ? (value ? 'toggle__knob--disabled--active' : 'toggle__knob--disabled') : ''
                }`}
            />
        </button>
    )
}
