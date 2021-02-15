import classNames from 'classnames'
import React, { useCallback } from 'react'

interface Icon {
    name: string
    value: number
    icon: JSX.Element
}

interface Props {
    icons: Icon[]
    disabled: boolean
    name: string
    selected?: number
    role: string
    onChange: (value: number) => void
}

export const IconRadioButtons: React.FunctionComponent<Props> = ({
    name,
    icons,
    selected,
    onChange,
    disabled,
    role,
}) => {
    const handleChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => onChange(Number(event.target.value)),
        [onChange]
    )

    return (
        <ul className="icon-radio-buttons" role={role}>
            {Object.values(icons).map(({ icon, name: iconName, value }) => (
                <li key={iconName} className="d-flex">
                    <label
                        className={classNames(
                            {
                                'icon-radio-buttons__label--inactive': selected !== undefined && value !== selected,
                                'icon-radio-buttons__label--active': value === selected,
                            },
                            'icon-radio-buttons__label'
                        )}
                    >
                        <input
                            disabled={disabled}
                            type="radio"
                            name={name}
                            onChange={handleChange}
                            value={value}
                            checked={value === selected}
                            aria-label={iconName}
                            className="icon-radio-buttons__input"
                        />
                        {icon}
                    </label>
                </li>
            ))}
        </ul>
    )
}
