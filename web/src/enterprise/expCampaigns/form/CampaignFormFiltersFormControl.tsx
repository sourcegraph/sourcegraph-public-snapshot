import React, { useCallback } from 'react'

interface Props {
    value: string
    onChange: (value: string) => void
    disabled?: boolean
    className?: string
}

export const CampaignFormFiltersFormControl: React.FunctionComponent<Props> = ({
    value,
    onChange: parentOnChange,
    disabled,
    className = '',
}) => {
    const onChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => parentOnChange(e.currentTarget.value),
        [parentOnChange]
    )

    return (
        <div className={`form-group ${className}`}>
            <label htmlFor="campaign-form__filters">Filters</label>
            <input
                type="text"
                id="campaign-form__filters"
                className="form-control"
                placeholder="repo:myrepo owner:alice"
                value={value}
                onChange={onChange}
                disabled={disabled}
            />
        </div>
    )
}
