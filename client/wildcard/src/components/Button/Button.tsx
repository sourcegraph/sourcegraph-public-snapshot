import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

import styles from './Button.module.scss'
import { BUTTON_VARIANTS, BUTTON_SIZES } from './constants'
import { getButtonSize, getButtonStyle } from './utils'

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
    /**
     * A tooltip to display when the user hovers the button.
     */
    ['data-tooltip']?: string
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
            outline,
            className,
            disabled,
            ...attributes
        },
        reference
    ) => {
        const tooltip = attributes['data-tooltip']

        const buttonComponent = (
            <Component
                ref={reference}
                className={classNames(
                    'btn',
                    variant && getButtonStyle({ variant, outline }),
                    size && getButtonSize({ size }),
                    className
                )}
                type={Component === 'button' ? type : undefined}
                disabled={disabled}
                {...attributes}
            >
                {children}
            </Component>
        )

        // Disabled elements don't fire mouse events, but the `Tooltip` relies on mouse
        // events. This restores the tooltip behavior for disabled buttons by rendering an
        // invisible `div` with the tooltip on top of the button, in the case that it is
        // disabled. See https://stackoverflow.com/a/3100395 for more.
        return disabled && tooltip ? (
            <div className={styles.container}>
                {disabled && tooltip ? <div className={styles.disabledTooltip} data-tooltip={tooltip} /> : null}
                {buttonComponent}
            </div>
        ) : (
            buttonComponent
        )
    }
) as ForwardReferenceComponent<'button', ButtonProps>
