import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

import styles from './Button.module.scss'
import { BUTTON_VARIANTS, BUTTON_SIZES, BUTTON_DISPLAY } from './constants'
import { getButtonSize, getButtonStyle, getButtonDisplay } from './utils'

export interface ButtonProps
    extends React.ButtonHTMLAttributes<HTMLButtonElement>,
        React.RefAttributes<HTMLButtonElement> {
    /**
     * The variant style of the button. Defaults to `primary`
     */
    variant?: typeof BUTTON_VARIANTS[number]
    /**
     * Allows modifying the size of the button. Supports larger or smaller variants.
     */
    size?: typeof BUTTON_SIZES[number]
    /**
     * Allows modifying the display property of the button. Supports inline-block or block variants.
     */
    display?: typeof BUTTON_DISPLAY[number]
    /**
     * Modifies the button style to have a transparent/light background and a more pronounced outline.
     */
    outline?: boolean
    /**
     * Used to change the element that is rendered.
     * Useful if needing to style a link as a button, or in certain cases where a different element is required.
     * Always be mindful of potentially accessibiliy pitfalls when using this!
     * Note: This component assumes `HTMLButtonElement` types, providing a different component here will change the potential types that can be passed to this component.
     */
    as?: React.ElementType
}

/**
 * Simple button.
 *
 * Style can be configured using different button `variant`s.
 *
 * Buttons should be used to allow users to trigger specific actions on the page.
 * Always be mindful of how intent is signalled to the user when using buttons. We should consider the correct button `variant` for each action.
 *
 * Some examples:
 * - The main action a user should take on the page should usually be styled with the `primary` variant.
 * - Other additional actions on the page should usually be styled with the `secondary` variant.
 * - A destructive 'delete' action should be styled with the `danger` variant.
 *
 * Tips:
 * - Avoid using button styling for links where possible. Buttons should typically trigger an action, links should navigate to places.
 */
export const Button = React.forwardRef(
    (
        {
            children,
            as: Component = 'button',
            type = 'button',
            variant,
            size,
            display,
            outline,
            className,
            ...attributes
        },
        reference
    ) => (
        <Component
            ref={reference}
            className={classNames(
                styles.btn,
                variant && getButtonStyle({ variant, outline }),
                size && getButtonSize({ size }),
                display && getButtonDisplay({ display }),
                className
            )}
            type={Component === 'button' ? type : undefined}
            {...attributes}
        >
            {children}
        </Component>
    )
) as ForwardReferenceComponent<'button', ButtonProps>
