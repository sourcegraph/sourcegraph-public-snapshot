import * as React from 'react'

import classNames from 'classnames'
import Check from 'mdi-react/CheckIcon'

import styles from './ToggleBig.module.scss'

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
    'data-testid'?: string
}

/** A big toggle switch input component. */
export const ToggleBig: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    disabled,
    className,
    id,
    title,
    value,
    tabIndex,
    onToggle,
    onHover,
    onFocus,
    'data-testid': dataTestId,
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
            className={classNames(styles.toggleBig, styles.container, className)}
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
            data-testid={dataTestId}
        >
            <span
                className={classNames(styles.bar, {
                    [styles.barOn]: value,
                })}
            />
            <span
                className={classNames(styles.barShadow, {
                    [styles.barShadowOn]: value,
                })}
            />
            <span
                className={classNames('d-flex flex-column justify-content-center align-items-center', styles.knob, {
                    [styles.knobOn]: value,
                })}
            >
                {value && <Check size={16} className={styles.iconOn} />}
            </span>
            <span className={classNames(styles.text, { [styles.textOn]: value })}>
                {value ? 'Enabled' : 'Disabled'}
            </span>
        </button>
    )
}
