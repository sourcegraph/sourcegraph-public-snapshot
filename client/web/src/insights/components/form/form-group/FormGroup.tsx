import React, { PropsWithChildren, Ref } from 'react'

interface FormGroupProps {
    /** Name attr value for root fieldset element. */
    name: string
    /** Title on top of group. */
    title: string
    /** Subtitle of group. */
    subtitle?: string
    /** Error message for field group. */
    error?: string
    /** Description text, renders below of content inputs of group. */
    description?: string
    /** Custom class name for root fieldset element. */
    className?: string
    /** Custom class name for div children wrapper element. */
    contentClassName?: string
    /** Reference to root fieldset element.*/
    innerRef?: Ref<HTMLFieldSetElement>
}

/** Displays fieldset (group) of fields for code insight creation form with error message. */
export const FormGroup: React.FunctionComponent<PropsWithChildren<FormGroupProps>> = props => {
    const { innerRef, className, contentClassName, name, title, subtitle, children, description, error } = props

    return (
        <fieldset ref={innerRef} name={name} className={className}>
            <legend className="d-flex flex-column mb-3">
                <div className="font-weight-bold">{title}</div>

                {/* Since safari doesn't support flex column on legend element we have to set d-block*/}
                {/* explicitly */}
                {subtitle && <small className="d-block text-muted">{subtitle}</small>}
                {error && (
                    <small role="alert" className="d-block text-danger">
                        {error}
                    </small>
                )}
            </legend>

            <div className={contentClassName}>{children}</div>

            {description && <small className="d-block mt-3 text-muted">{description}</small>}
        </fieldset>
    )
}
