import classNames from 'classnames'
import React, { useCallback } from 'react'

import styles from './IconRadioButtons.module.scss'

interface Icon {
    name: string
    value: number
    icon: React.ComponentType
}

interface Props {
    icons: Icon[]
    disabled: boolean
    name: string
    selected?: number
    onChange: (value: number) => void
}

/**
 * Used to render a list of icons with <input type="radio" />
 */
export const IconRadioButtons: React.FunctionComponent<Props> = ({ name, icons, selected, onChange, disabled }) => {
    const handleChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => onChange(Number(event.target.value)),
        [onChange]
    )

    return (
        <ul className={styles.buttons}>
            {Object.values(icons).map(({ icon: Icon, name: iconName, value }) => (
                <li key={iconName} className="d-flex">
                    <label className={styles.label}>
                        <input
                            disabled={disabled}
                            type="radio"
                            name={name}
                            onChange={handleChange}
                            value={value}
                            checked={value === selected}
                            aria-label={iconName}
                            className={styles.input}
                        />
                        <span
                            className={classNames(styles.border, {
                                [styles.borderInactive]: selected !== undefined && value !== selected,
                                [styles.borderActive]: value === selected,
                            })}
                            aria-hidden={true}
                        />
                        <span
                            className={classNames(styles.emoji, {
                                [styles.emojiInactive]: selected !== undefined && value !== selected,
                                [styles.emojiActive]: value === selected,
                            })}
                        >
                            <Icon />
                        </span>
                    </label>
                </li>
            ))}
        </ul>
    )
}
