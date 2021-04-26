import classnames from 'classnames'
import React, { PropsWithChildren } from 'react'

import styles from './FormGroup.module.scss'

interface FormGroupProps {
    /** Name of form group. Title on top of group. */
    name: string
    /** Subtitle of group. */
    subtitle?: string
    /** Error message for field group. */
    error?: string
    /** Description text, renders below of content inputs of group. */
    description?: string
    /** Custom class name for root fieldset element. */
    className?: string
}

/** Displays fieldset (group) of fields for code insight creation form with error message. */
export const FormGroup: React.FunctionComponent<PropsWithChildren<FormGroupProps>> = props => {
    const { className, name, subtitle, children, description, error } = props

    return (
        <fieldset
            className={classnames(styles.formGroup, className, {
                [styles.formGroupWithSubtitle]: !!subtitle,
            })}
        >
            <div className={styles.formGroupNameBlock}>
                <h4 className={styles.formGroupName}>{name}</h4>

                {subtitle && <span className="text-muted">{subtitle}</span>}
                {error && <span className={styles.formGroupError}>*{error}</span>}
            </div>

            <div>{children}</div>

            {description && (
                <span className={classnames(styles.formGroupDescription, 'text-muted')}>{description}</span>
            )}
        </fieldset>
    )
}
