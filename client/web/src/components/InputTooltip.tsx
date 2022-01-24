import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import styles from './InputTooltip.module.scss'

export interface InputTooltipProps
    extends React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement> {
    tooltip: string
}

/**
 * A wrapper around `input` that restores the hover tooltip capability even if the input is disabled.
 *
 * Disabled elements do not trigger mouse events on hover, and `Tooltip` relies on mouse events.
 *
 * All other props are passed to the `input` element.
 */
export const InputTooltip: React.FunctionComponent<InputTooltipProps> = ({ disabled, tooltip, ...props }) => (
    <div className={styles.container}>
        {disabled ? <div className={styles.disabledTooltip} data-tooltip={tooltip} /> : null}
        <Button<'input', React.DetailedHTMLProps<React.InputHTMLAttributes<HTMLInputElement>, HTMLInputElement>>
            as="input"
            disabled={disabled}
            data-tooltip={disabled ? undefined : tooltip}
            {...props}
        />
    </div>
)
