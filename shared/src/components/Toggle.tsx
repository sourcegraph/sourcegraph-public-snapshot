import * as React from 'react'
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
}) => {
    function onClick(): void {
        if (!disabled && onToggle) {
            onToggle(!value)
        }
    }

    return (
        <button
            type="button"
            className={classnames('toggle', className, {})}
            id={id}
            title={title}
            value={value ? 1 : 0}
            onClick={onClick}
            tabIndex={tabIndex}
            disabled={disabled}
            role="switch"
            aria-checked={value}
        >
            <span
                className={classnames('toggle__bar', {
                    'toggle__bar--on': value,
                })}
            />
            <span
                className={classnames('toggle__bar-shadow', {
                    'toggle__bar-shadow--on': value,
                })}
            />
            <span
                className={classnames('toggle__knob', {
                    'toggle__knob--on': value,
                })}
            />
        </button>
    )
}
