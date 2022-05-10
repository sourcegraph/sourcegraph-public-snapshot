import React from 'react'

import { Button, ButtonProps } from '@sourcegraph/wildcard'

import styles from './InputTooltip.module.scss'

type ButtonAndInputElementProps = Omit<ButtonProps, 'type'> & React.InputHTMLAttributes<HTMLInputElement>

export interface InputTooltipProps extends ButtonAndInputElementProps {
    tooltip: string
}

/**
 * A wrapper around `input` that restores the hover tooltip capability even if the input is disabled.
 *
 * Disabled elements do not trigger mouse events on hover, and `Tooltip` relies on mouse events.
 *
 * All other props are passed to the `input` element.
 */
export const InputTooltip: React.FunctionComponent<React.PropsWithChildren<InputTooltipProps>> = ({
    disabled,
    tooltip,
    type,
    ...props
}) => (
    <div className={styles.container}>
        {disabled ? <div className={styles.disabledTooltip} data-tooltip={tooltip} /> : null}
        <Button
            as="input"
            disabled={disabled}
            className={disabled ? styles.disabledBtn : undefined}
            data-tooltip={disabled ? undefined : tooltip}
            type={type as ButtonProps['type']}
            {...props}
        />
    </div>
)
