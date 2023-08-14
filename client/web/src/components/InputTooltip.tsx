import React from 'react'

import { Button, type ButtonProps, Tooltip, type TooltipProps } from '@sourcegraph/wildcard'

import styles from './InputTooltip.module.scss'

type ButtonAndInputElementProps = Omit<ButtonProps, 'type'> & React.InputHTMLAttributes<HTMLInputElement>

export interface InputTooltipProps extends ButtonAndInputElementProps {
    tooltip: string
    placement?: TooltipProps['placement']
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
    placement,
    type,
    ...props
}) => (
    <div className={styles.container}>
        <Tooltip content={tooltip} placement={placement}>
            <Button as="input" disabled={disabled} type={type as ButtonProps['type']} {...props} />
        </Tooltip>
    </div>
)
