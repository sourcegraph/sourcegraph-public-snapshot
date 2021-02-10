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

export const IconRadioButtons: React.FunctionComponent<Props> = ({ name, icons, selected, onChange }) => {
    const handleChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => onChange(Number(event.target.value)),
        []
    )

    return (
        <ul className="icon-radio-buttons d-flex justify-content-around">
            {Object.values(icons).map(({ icon, name: iconName, value }) => (
                <li
                    key={iconName}
                    className={`icon-radio-buttons__button ${
                        value === selected ? 'icon-radio-buttons__button--active' : ''
                    }`}
                >
                    <input
                        type="radio"
                        name={name}
                        id={iconName}
                        onChange={handleChange}
                        value={value}
                        checked={value === selected}
                        aria-label={iconName}
                        className="icon-radio-buttons__input"
                    />
                    <label htmlFor={iconName} className="icon-radio-buttons__label">
                        {icon}
                    </label>
                </li>
            ))}
        </ul>
    )
}
