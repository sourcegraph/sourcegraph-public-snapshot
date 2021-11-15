import classNames from 'classnames'
import * as React from 'react'

import styles from './Toggle.module.scss'

interface Props {
    /** The initial value. */
    value?: boolean

    /** The DOM ID of the element. */
    id?: string

    /**
     * Called when the user changes the input's value.
     */
    onToggle?: (value: boolean) => void

    onClick?: (event: React.MouseEvent<HTMLButtonElement>) => void

    /** The title attribute (tooltip). */
    title?: string

    'aria-label'?: string
    'aria-labelledby'?: string
    'aria-describedby'?: string
    'data-testid'?: string

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
    onClick,
    'aria-label': ariaLabel,
    'aria-labelledby': ariaLabelledby,
    'aria-describedby': ariaDescribedby,
    'data-testid': dataTestId,
}) => {
    function onButtonClick(event: React.MouseEvent<HTMLButtonElement>): void {
        event.stopPropagation()
        if (!disabled && onToggle) {
            onToggle(!value)
        }
        if (!disabled && onClick) {
            onClick(event)
        }
    }

    return (
        <button
            type="button"
            className={classNames(styles.toggle, className)}
            id={id}
            title={title}
            value={value ? 1 : 0}
            onClick={onButtonClick}
            tabIndex={tabIndex}
            disabled={disabled}
            role="switch"
            aria-checked={value}
            aria-label={ariaLabel}
            aria-labelledby={ariaLabelledby}
            aria-describedby={ariaDescribedby}
            data-testid={dataTestId}
        >
            <span
                className={classNames(styles.bar, {
                    [styles.barOn]: value,
                })}
            />
            <span
                className={classNames(styles.knob, {
                    [styles.knobOn]: value,
                })}
            />
        </button>
    )
}
