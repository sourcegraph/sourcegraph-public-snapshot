import classnames from 'classnames'
import React, { InputHTMLAttributes } from 'react'

interface RadioInputProps extends InputHTMLAttributes<HTMLInputElement> {
    /** Title of radio input. */
    title: string
    /** Description text for radio input. */
    description?: string
    /** Custom class name for root label element. */
    className?: string

    labelTooltipText?: string
}

/** Displays form radio input for code insight creation form. */
export const FormRadioInput: React.FunctionComponent<RadioInputProps> = props => {
    const { title, description, className, labelTooltipText, ...otherProps } = props

    return (
        <label data-tooltip={labelTooltipText} className={classnames('d-flex align-items-center', className)}>
            <input type="radio" {...otherProps} />

            <div className="pl-2">
                <span>{title}</span>
                {description && <span className="text-muted"> – {description}</span>}
            </div>
        </label>
    )
}
