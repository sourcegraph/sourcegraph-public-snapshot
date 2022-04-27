import { forwardRef } from 'react'

import { ForwardReferenceComponent } from '../../../types'
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
export const FormFieldLabel = forwardRef(({ htmlFor, className, children, ...rest }, reference) => (
    <Label htmlFor={htmlFor} className={className} ref={reference} {...rest}>
        {children}
    </Label>
)) as ForwardReferenceComponent<'label', FormFieldLabelProps>
