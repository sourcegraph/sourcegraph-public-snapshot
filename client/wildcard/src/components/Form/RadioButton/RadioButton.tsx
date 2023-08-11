import React from 'react'

import { BaseControlInput, type ControlInputProps } from '../internal/BaseControlInput'

export type RadioButtonProps = {
    /**
     * The name of the radio group. Used to group radio controls together to ensure mutual exclusivity.
     * If you do not need this prop, consider if a checkbox is better suited for your use case.
     */
    name: string
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
export const RadioButton: React.FunctionComponent<React.PropsWithChildren<RadioButtonProps>> = React.forwardRef(
    (props, reference) => <BaseControlInput {...props} type="radio" ref={reference} />
)
