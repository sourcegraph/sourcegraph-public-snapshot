import classnames from 'classnames'
import React, { InputHTMLAttributes } from 'react'

import styles from './FormRadioInput.module.scss'

interface RadioInputProps extends InputHTMLAttributes<HTMLInputElement> {
    /** Title of radio input. */
    title: string
    /** Description text for radio input. */
    description?: string
    /** Custom class name for root label element. */
    className?: string
}

/** Displays form radio input for code insight creation form. */
export const FormRadioInput: React.FunctionComponent<RadioInputProps> = props => {
    const { title, description, className, ...otherProps } = props

    return (
        <label className={classnames(styles.radioInput, className)}>
            <input type="radio" className={classnames(styles.radioInputInput, 'form-control')} {...otherProps} />

            <div className={classnames(styles.radioInputDescriptionContent)}>
                <span>{title}</span>
                {description && <span className="text-muted"> – {description}</span>}
            </div>
        </label>
    )
}
