import classNames from 'classnames'
import React from 'react'

export interface SelectProps
    extends React.SelectHTMLAttributes<HTMLSelectElement>,
        React.RefAttributes<HTMLSelectElement> {
    className?: string
    isValid?: boolean
    label: React.ReactNode
    message?: React.ReactNode
}

export const Select: React.FunctionComponent<SelectProps> = React.forwardRef(
    ({ children, className, label, message, isValid, ...selectProps }, reference) => (
        <div className="form-check">
            <label className="form-check-label">
                <select ref={reference} className={classNames('form-control', className)} {...selectProps}>
                    {children}
                </select>
                {label}
            </label>
            {message && <small className={'field-message'}>{message}</small>}
        </div>
    )
)
