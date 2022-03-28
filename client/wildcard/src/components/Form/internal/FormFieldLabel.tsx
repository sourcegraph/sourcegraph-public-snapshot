import React from 'react'

import { Label } from '../../Typography/Label'

export interface FormFieldLabelProps {
    /**
     * The id of the form field to associate with the label.
     */
    htmlFor: string
    className?: string
}

/**
 * A simple label to render alongside a form field.
 */
export const FormFieldLabel: React.FunctionComponent<FormFieldLabelProps> = ({ htmlFor, className, children }) => (
    <Label htmlFor={htmlFor} className={className}>
        {children}
    </Label>
)
