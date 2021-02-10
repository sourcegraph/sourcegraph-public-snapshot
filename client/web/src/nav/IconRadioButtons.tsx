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
    onChange: (value: number) => void
}

export const IconRadioButtons: React.FunctionComponent<Props> = ({ name, icons, selected, onChange, disabled }) => {
    const handleChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => onChange(Number(event.target.value)),
        []
    )

    return (
        <ul className="icon-radio-buttons d-flex justify-content-around">
            {Object.values(icons).map(({ icon, name: iconName, value }) => (
                <li key={iconName} className="d-flex">
                    <label
                        htmlFor={iconName}
                        className={classNames(
                            {
                                'icon-radio-buttons__label--inactive': selected !== undefined && value !== selected,
                                'icon-radio-buttons__label--active': value === selected,
                            },
                            'icon-radio-buttons__label d-flex justify-content-center align-items-center'
                        )}
                    >
                        <input
                            disabled={disabled}
                            type="radio"
                            name={name}
                            id={iconName}
                            onChange={handleChange}
                            value={value}
                            checked={value === selected}
                            aria-label={iconName}
                            className="icon-radio-buttons__label--input"
                        />
                        {icon}
                    </label>
                </li>
            ))}
        </ul>
    )
}
