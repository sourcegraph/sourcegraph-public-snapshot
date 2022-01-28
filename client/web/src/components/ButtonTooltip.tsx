import React from 'react'

import { Button, ButtonProps } from '@sourcegraph/wildcard'

import styles from './InputTooltip.module.scss'

interface ButtonTooltipProps extends ButtonProps {
    tooltip?: string
}

/**
 * A wrapper around `button` that restores the hover tooltip capability even if the button is disabled.
 *
 * Disabled elements do not trigger mouse events on hover, and `Tooltip` relies on mouse events.
 *
 * All other props are passed to the `button` element.
 */
export const ButtonTooltip: React.FunctionComponent<ButtonTooltipProps> = ({ disabled, tooltip, ...props }) => (
    <div className={styles.container}>
        {disabled && tooltip ? <div className={styles.disabledTooltip} data-tooltip={tooltip} /> : null}
        <Button disabled={disabled} data-tooltip={disabled ? undefined : tooltip} {...props} />
    </div>
)
