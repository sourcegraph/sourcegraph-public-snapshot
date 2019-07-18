import React from 'react'

/**
 * A form input for a rule's name.
 */
export const RuleNameFormGroup: React.FunctionComponent<{
    className?: string
    value?: string
    disabled?: boolean
    onChange: React.ChangeEventHandler<HTMLInputElement>
}> = ({ className, value = '', disabled, onChange }) => (
    <div className={`form-group ${className}`}>
        <label htmlFor="rule-name-form-group__input">Name</label>
        <input
            type="text"
            name="rule-name"
            className="form-control"
            id="rule-name-form-group__input"
            onChange={onChange}
            required={true}
            autoComplete="off"
            autoFocus={true}
            value={value}
            disabled={disabled}
        />
    </div>
)
