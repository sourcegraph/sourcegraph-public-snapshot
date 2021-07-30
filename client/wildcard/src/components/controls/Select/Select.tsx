import classNames from 'classnames'
import React from 'react'

export interface SelectProps
    extends React.SelectHTMLAttributes<HTMLSelectElement>,
        React.RefAttributes<HTMLSelectElement>,
        Omit<FormFieldLabelProps, 'children' | 'className'> {
    className?: string
    /**
     * Used to control the styling of the <select> and surrounding elements.
     * Set this value to `false` to show invalid styling.
     * Set this value to `true` to show valid styling.
     */
    isValid?: boolean
    /**
     * Optional message to display below the <select>.
     * This should typically be used to display additional information (perhap error/success states) to the user.
     * It will be styled differently if `isValid` is truthy or falsy.
     */
    message?: React.ReactNode
    /**
     * Use the Bootstrap custom <select> styles
     */
    isCustomStyle?: boolean
}

export const getSelectStyles = (isCustomStyle?: boolean): string => {
    if (isCustomStyle) {
        return 'custom-select'
    }

    return 'form-control'
}

export const Select: React.FunctionComponent<SelectProps> = React.forwardRef(
    ({ children, className, label, displayLabel, message, isValid, isCustomStyle, ...selectProps }, reference) => (
        <div className="form-group">
            <FormFieldLabel label={label} displayLabel={displayLabel}>
                <select
                    ref={reference}
                    className={classNames(getSelectStyles(isCustomStyle), className)}
                    {...selectProps}
                >
                    {children}
                </select>
            </FormFieldLabel>
            {message && <small className="field-message">{message}</small>}
        </div>
    )
)

interface FormFieldLabelProps {
    displayLabel?: boolean
    /**
     * Descriptive text for the form control.
     * Uses <label> if `displayLabel` is `true`.
     * Uses 'aria-label` if `displayLabel` is `false`.
     */
    label: string
    /**
     * Styles to apply to the <label>
     */
    className?: string

    children: React.ReactElement
}

const FormFieldLabel: React.FunctionComponent<FormFieldLabelProps> = ({ children, label, displayLabel, className }) => {
    if (displayLabel) {
        return (
            <label className={className}>
                {children}
                {label}
            </label>
        )
    }

    return React.cloneElement(children, { 'aria-label': label })
}
