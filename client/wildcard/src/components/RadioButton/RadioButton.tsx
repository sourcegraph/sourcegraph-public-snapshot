import classNames from 'classnames'
import React from 'react'
import {} from 'reactstrap'

export interface RadioButtonProps
    extends React.InputHTMLAttributes<HTMLInputElement>,
        React.RefAttributes<HTMLInputElement> {
    className?: string
    label: string
    valid?: boolean
    message?: string
    checked?: boolean
}

const getValidStyle = (valid?: RadioButtonProps['valid']): string => {
    if (valid === undefined) {
        return ''
    }

    if (valid) {
        return 'is-valid'
    }

    return 'is-invalid'
}

const getMessageStyle = (valid?: RadioButtonProps['valid']): string => {
    if (valid === undefined) {
        return 'field-message'
    }

    if (valid) {
        return 'valid-feedback'
    }

    return 'invalid-feedback'
}

export const RadioButton: React.FunctionComponent<RadioButtonProps> = React.forwardRef(
    ({ children, label, message, valid, ...inputProps }, reference) => (
        <div className="form-check">
            <label className="form-check-label">
                <input
                    ref={reference}
                    type="radio"
                    className={classNames('form-check-input', getValidStyle(valid))}
                    {...inputProps}
                />
                {label}
            </label>
            {message && <small className={getMessageStyle(valid)}>{message}</small>}
        </div>
    )
)
