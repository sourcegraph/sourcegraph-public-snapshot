import classNames from 'classnames'
import React from 'react'

import { ControlInputProps } from '../internal/BaseControlInput'
import { getValidStyle } from '../internal/utils'

export type RadioButtonProps = {
    /**
     * The name of the radio group. Used to group radio controls together to ensure mutual exclusivity.
     * If you do not need this prop, consider if a checkbox is better suited for your use case.
     */
    name: string
    labelProps: React.LabelHTMLAttributes<HTMLLabelElement>
} & ControlInputProps

/**
 * Renders a single radio button.
 *
 * Radio buttons should be used when a user must make a single choice from a list of two or more mutually exclusive options.
 *
 * Grouped radio buttons should be visually presented together.
 *
 * Useful article comparing radio buttons to checkboxes: https://www.nngroup.com/articles/checkboxes-vs-radio-buttons/
 */
export const RadioButton: React.FunctionComponent<RadioButtonProps> = React.forwardRef(
    ({ isValid, labelProps, children, ...props }, reference) => {
        const isValidStyle = getValidStyle(isValid)

        return (
            <label {...labelProps} className={classNames('form-check-label', isValidStyle, labelProps.className)}>
                <input
                    {...props}
                    ref={reference}
                    type="radio"
                    className={classNames('form-check-input', isValidStyle, props.className)}
                />
                {children}
            </label>
        )
    }
)

RadioButton.displayName = 'RadioButton'
