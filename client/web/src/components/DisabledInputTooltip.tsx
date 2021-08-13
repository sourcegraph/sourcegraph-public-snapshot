import React from 'react'

import styles from './DisabledInputTooltip.module.scss'

/**
 * A wrapper around `input` that restores the hover tooltip capability even if the input is disabled.
 *
 * Disabled elements do not trigger mouse events on hover, and `Tooltip` relies on mouse events.
 *
 * All other props are passed to the `input` element.
 */
export const DisabledInputTooltip: React.FunctionComponent<
    React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> & { tooltip: string }
> = ({ tooltip, ...props }) => (
    <div className={styles.container}>
        <div className={styles.containerTooltip} data-tooltip={tooltip} />
        <input {...props} />
    </div>
)
