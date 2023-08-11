import { forwardRef } from 'react'

import { Label } from '../..'
import type { ForwardReferenceComponent } from '../../../types'

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
export const FormFieldLabel = forwardRef(function FormFieldLabel({ htmlFor, className, children, ...rest }, reference) {
    return (
        <Label htmlFor={htmlFor} className={className} ref={reference} {...rest}>
            {children}
        </Label>
    )
}) as ForwardReferenceComponent<'label', FormFieldLabelProps>
