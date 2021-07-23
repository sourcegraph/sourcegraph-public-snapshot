import classNames from 'classnames'
import React from 'react'
import { BUTTON_VARIANTS, BUTTON_SIZES } from './constants'
import { getButtonSize, getButtonStyle } from './utils'

export interface ButtonProps
    extends React.ButtonHTMLAttributes<HTMLButtonElement>,
        React.RefAttributes<HTMLButtonElement> {
    variant?: typeof BUTTON_VARIANTS[number]
    size?: typeof BUTTON_SIZES[number]
    outline?: boolean
    as?: React.ElementType
}

export const Button: React.FunctionComponent<ButtonProps> = React.forwardRef(
    (
        {
            children,
            as: Component = 'button',
            type = 'button',
            variant = 'primary',
            size = 'md',
            outline,
            className,
            ...attributes
        },
        ref
    ) => {
        return (
            <Component
                ref={ref}
                className={classNames('btn', getButtonStyle({ variant, outline }), getButtonSize({ size }), className)}
                type={Component === 'button' ? type : undefined}
                {...attributes}
            >
                {children}
            </Component>
        )
    }
)
