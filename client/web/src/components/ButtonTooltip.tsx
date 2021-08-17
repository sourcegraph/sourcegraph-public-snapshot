import React from 'react'

import styles from './InputTooltip.module.scss'

type ButtonProps = React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement>

/**
 * A wrapper around `button` that restores the hover tooltip capability even if the button is disabled.
 *
 * Disabled elements do not trigger mouse events on hover, and `Tooltip` relies on mouse events.
 *
 * All other props are passed to the `button` element.
 */
export const ButtonTooltip: React.FunctionComponent<
    ButtonProps & { tooltip?: string } & Required<Pick<ButtonProps, 'type'>>
> = ({ disabled, tooltip, type, ...props }) => (
    <div className={styles.container}>
        {disabled && tooltip ? <div className={styles.containerTooltip} data-tooltip={tooltip} /> : null}
        {/* eslint-disable-next-line react/button-has-type */}
        <button type={type} disabled={disabled} data-tooltip={disabled ? undefined : tooltip} {...props} />
    </div>
)
