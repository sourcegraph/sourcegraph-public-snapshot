import classnames from 'classnames'
import Check from 'mdi-react/CheckIcon'
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

    /**
     * Called when the user focuses on the toggle.
     */
    onFocus?: (value: boolean) => void

    /** The title attribute (tooltip). */
    title?: string

    disabled?: boolean
    tabIndex?: number
    className?: string

    /** Data attribute for testing */
    dataTest?: string
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
    onFocus,
    dataTest,
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

    function onToggleFocus(): void {
        if (onFocus) {
            onFocus(!value)
        }
    }

    return (
        <button
            type="button"
            className={classnames('toggle-big toggle-big__container', className)}
            id={id}
            title={title}
            value={value ? 1 : 0}
            onClick={onClick}
            tabIndex={tabIndex}
            onMouseOver={onMouseOver}
            onFocus={onToggleFocus}
            disabled={!!disabled}
            role="switch"
            aria-checked={value}
            data-test={dataTest}
        >
            <span
                className={classnames('toggle-big__bar', {
                    'toggle-big__bar--on': value,
                })}
            />
            <span
                className={classnames('toggle-big__bar-shadow', {
                    'toggle-big__bar-shadow--on': value,
                })}
            />
            <span
                className={classnames('toggle-big__knob d-flex flex-column justify-content-center align-items-center', {
                    'toggle-big__knob--on': value,
                })}
            >
                {value && <Check size={16} className="toggle-big__icon--on" />}
            </span>
            <span className={classnames('toggle-big__text', { 'toggle-big__text--on': value })}>
                {value ? 'Enabled' : 'Disabled'}
            </span>
        </button>
    )
}
