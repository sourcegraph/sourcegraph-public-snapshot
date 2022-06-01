import { forwardRef } from 'react'

import { Typography } from '../..'
import { ForwardReferenceComponent } from '../../../types'

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
    <Typography.Label htmlFor={htmlFor} className={className} ref={reference} {...rest}>
        {children}
    </Typography.Label>
)) as ForwardReferenceComponent<'label', FormFieldLabelProps>
