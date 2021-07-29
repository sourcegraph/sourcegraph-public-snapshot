import classNames from 'classnames'
import React from 'react'

import { getMessageStyle, getValidStyle } from '../utils'

export interface CheckboxProps
    extends React.InputHTMLAttributes<HTMLInputElement>,
        React.RefAttributes<HTMLInputElement> {
    className?: string
    isValid?: boolean
    label: React.ReactNode
    message?: React.ReactNode
}

export const Checkbox: React.FunctionComponent<CheckboxProps> = React.forwardRef(
    ({ children, label, message, isValid, ...inputProps }, reference) => (
        <div className="form-check">
            <label className="form-check-label">
                <input
                    ref={reference}
                    type="checkbox"
                    className={classNames('form-check-input', getValidStyle(isValid))}
                    {...inputProps}
                />
                {label}
            </label>
            {message && <small className={getMessageStyle(isValid)}>{message}</small>}
        </div>
    )
)
