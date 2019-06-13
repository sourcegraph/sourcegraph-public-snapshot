import React from 'react'

/**
 * A form input for a thread's title.
 */
export const ThreadTitleFormGroup: React.FunctionComponent<{
    className?: string
    value?: string
    disabled?: boolean
    onChange: React.ChangeEventHandler<HTMLInputElement>
}> = ({ className, value = '', disabled, onChange }) => (
    <div className={`form-group ${className}`}>
        <label htmlFor="thread-title-form-group__input">Title</label>
        <input
            type="text"
            name="thread-title"
            className="form-control"
            id="thread-title-form-group__input"
            onChange={onChange}
            required={true}
            autoComplete="off"
            autoFocus={true}
            value={value}
            disabled={disabled}
        />
    </div>
)
