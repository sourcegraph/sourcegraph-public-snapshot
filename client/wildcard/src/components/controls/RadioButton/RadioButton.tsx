import classNames from 'classnames'
import React from 'react'

import { getMessageStyle, getValidStyle } from '../utils'

export interface RadioButtonProps
    extends React.InputHTMLAttributes<HTMLInputElement>,
        React.RefAttributes<HTMLInputElement> {
    className?: string
    isValid?: boolean
    label: React.ReactNode
    message?: React.ReactNode
}

export const RadioButton: React.FunctionComponent<RadioButtonProps> = React.forwardRef(
    ({ children, label, message, isValid, ...inputProps }, reference) => (
        <div className="form-check">
            <label className="form-check-label">
                <input
                    ref={reference}
                    type="radio"
                    className={classNames('form-check-input', getValidStyle(isValid))}
                    {...inputProps}
                />
                {label}
            </label>
            {message && <small className={getMessageStyle(isValid)}>{message}</small>}
        </div>
    )
)
