import React, { useCallback } from 'react'

interface Props {
    value: string
    onChange: (newValue: string) => void
    autoFocus?: boolean
}

/**
 * The title field for a thread in a form.
 */
export const ThreadFormTitleField: React.FunctionComponent<Props> = ({ value, onChange, autoFocus }) => {
    const onTitleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => onChange(e.currentTarget.value),
        [onChange]
    )
    return (
        <div className="form-group">
            <label htmlFor="thread-form__title">Title</label>
            <input
                type="text"
                id="thread-form__title"
                className="form-control"
                required={true}
                minLength={1}
                placeholder="Title"
                value={value}
                onChange={onTitleChange}
                autoFocus={autoFocus}
            />
        </div>
    )
}
