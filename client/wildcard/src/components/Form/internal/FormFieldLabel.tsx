import React from 'react'

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
    <label htmlFor={htmlFor} className={className}>
        {children}
    </label>
)
