import React from 'react'

import classNames from 'classnames'

import { Label } from '../../../Typography'
import { InputProps } from '../Input'

export interface InputLabelProps extends Pick<InputProps, 'variant'> {
    /** Text label of input. */
    label?: React.ReactNode

    /** Custom class name for root label element. */
    className?: string
}

/**
 * A styled label to render around an input field.
 */
export const InputLabel: React.FunctionComponent<React.PropsWithChildren<InputLabelProps>> = props => {
    const { label, variant = 'regular', className, children, ...rest } = props

    return (
        <Label className={classNames('w-100', className)} {...rest}>
            {label && <div className="mb-2">{variant === 'regular' ? label : <small>{label}</small>}</div>}
            {children}
        </Label>
    )
}
