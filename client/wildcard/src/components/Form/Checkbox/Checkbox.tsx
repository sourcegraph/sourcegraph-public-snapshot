import React from 'react'

import { BaseControlInput, type ControlInputProps } from '../internal/BaseControlInput'

export type CheckboxProps = ControlInputProps

/**
 * Renders a single checkbox.
 *
 * Checkboxes should be used when a user can select any number of choices from a list of options.
 * They can often be used stand-alone, for a single option that a user can turn on or off.
 *
 * Grouped checkboxes should be visually presented together.
 *
 * Useful article comparing checkboxes to radio buttons: https://www.nngroup.com/articles/checkboxes-vs-radio-buttons/
 */
export const Checkbox: React.FunctionComponent<React.PropsWithChildren<CheckboxProps>> = React.forwardRef(
    function Checkbox(props, reference) {
        return <BaseControlInput {...props} type="checkbox" ref={reference} />
    }
)
