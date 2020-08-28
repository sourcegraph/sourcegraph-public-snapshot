import * as React from 'react'
import Close from 'mdi-react/CloseIcon'
import Check from 'mdi-react/CheckIcon'
import classnames from 'classnames'

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

/** A big toggle switch input component. */
export const ToggleBig: React.FunctionComponent<Props> = ({
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
            className={classnames('toggle-big', className, {
                'toggle-big--disabled': !!disabled,
            })}
            id={id}
            title={title}
            value={value ? 1 : 0}
            onClick={onClick}
            tabIndex={tabIndex}
            onMouseOver={onMouseOver}
            disabled={!!disabled}
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
