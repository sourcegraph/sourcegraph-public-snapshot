import classnames from 'classnames'
import React, { PropsWithChildren } from 'react'

import styles from './FormGroup.module.scss'

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
}

/** Displays fieldset (group) of fields for code insight creation form with error message. */
export const FormGroup: React.FunctionComponent<PropsWithChildren<FormGroupProps>> = props => {
    const { className, contentClassName, name, title, subtitle, children, description, error } = props

    return (
        <fieldset name={name} className={className}>
            <legend className={classnames(styles.formGroupLegend, 'd-flex flex-column')}>
                <div className="mb-1 font-weight-bold">{title}</div>

                {subtitle && <small className="text-muted">{subtitle}</small>}
                {error && <small role='alert' className="text-danger">{error}</small>}
            </legend>

            <div className={contentClassName}>{children}</div>

            {description && <small className="d-block mt-3 text-muted">{description}</small>}
        </fieldset>
    )
}
