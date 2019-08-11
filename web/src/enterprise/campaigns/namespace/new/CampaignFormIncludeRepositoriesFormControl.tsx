import React, { useCallback } from 'react'

interface Props {
    value: string
    onChange: (value: string) => void
    disabled?: boolean
    className?: string
}

export const CampaignFormIncludeRepositoriesFormControl: React.FunctionComponent<Props> = ({
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
            <label htmlFor="campaign-form-common-fields__includeRepositories">Include only specific repositories</label>
            <input
                type="text"
                id="campaign-form-common-fields__includeRepositories"
                className="form-control"
                placeholder="Regular expression (e.g., myorg/)"
                value={value}
                onChange={onChange}
                disabled={disabled}
            />
        </div>
    )
}
