import classNames from 'classnames'
import React from 'react'

import { getMessageStyle, getValidStyle } from './utils'

export interface BaseControlInputProps
    extends React.InputHTMLAttributes<HTMLInputElement>,
        React.RefAttributes<HTMLInputElement> {
    className?: string
    isValid?: boolean
    label: React.ReactNode
    message?: React.ReactNode
}

export const BaseControlInput: React.FunctionComponent<BaseControlInputProps> = React.forwardRef(
    ({ children, label, message, isValid, type, ...inputProps }, reference) => (
        <div className="form-check">
            <label className="form-check-label">
                <input
                    ref={reference}
                    type={type}
                    className={classNames('form-check-input', getValidStyle(isValid))}
                    {...inputProps}
                />
                {label}
            </label>
            {message && <small className={getMessageStyle(isValid)}>{message}</small>}
        </div>
    )
)
