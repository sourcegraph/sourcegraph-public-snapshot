import React from 'react'

import styles from './InputTooltip.module.scss'

/**
 * A wrapper around `input` that restores the hover tooltip capability even if the input is disabled.
 *
 * Disabled elements do not trigger mouse events on hover, and `Tooltip` relies on mouse events.
 *
 * All other props are passed to the `input` element.
 */
export const InputTooltip: React.FunctionComponent<
    React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> & { tooltip: string }
> = ({ disabled, tooltip, ...props }) => (
    <div className={styles.container}>
        {disabled ? <div className={styles.tooltip} data-tooltip={tooltip} /> : null}
        <input disabled={disabled} data-tooltip={disabled ? undefined : tooltip} {...props} />
    </div>
)
