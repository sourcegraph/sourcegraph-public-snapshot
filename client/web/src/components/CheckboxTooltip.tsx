import { FC } from 'react'

import classNames from 'classnames'

import { Checkbox, CheckboxProps, Tooltip, TooltipProps } from '@sourcegraph/wildcard'

import styles from './InputTooltip.module.scss'

export type CheckboxTooltipProps = CheckboxProps & {
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
export const CheckboxTooltip: FC<CheckboxTooltipProps> = ({ tooltip, placement, className, ...props }) => (
    <Tooltip content={tooltip} placement={placement}>
        <Checkbox
            {...props}
            wrapperClassName={styles.checkboxWrapper}
            className={classNames(className, styles.checkbox)}
        />
    </Tooltip>
)
