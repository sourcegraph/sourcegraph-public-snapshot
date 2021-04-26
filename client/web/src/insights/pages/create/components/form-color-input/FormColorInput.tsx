import classnames from 'classnames'
import React, { ChangeEvent, ChangeEventHandler, useCallback, useRef, useState } from 'react'
import { noop } from 'rxjs'

import styles from './FormColorInput.module.scss'

interface FormColorPickerProps {
    /** Name of data series color */
    name?: string
    /** Value of data series color ()*/
    value?: string
    /** Different values of preset color. */
    colours?: { color: string; name?: string }[]
    /** Change listener to track changes of color radio group. */
    onChange?: ChangeEventHandler<HTMLInputElement>
    /** Custom class name. */
    className?: string
}

const DEFAULT_COLOURS = [
    { color: 'var(--oc-red-7)', name: 'Red' },
    { color: 'var(--oc-pink-7)', name: 'Pink' },
    { color: 'var(--oc-grape-7)', name: 'Grape' },
    { color: 'var(--oc-violet-7)', name: 'Violet' },
    { color: 'var(--oc-indigo-7)', name: 'Indigo' },
    { color: 'var(--oc-blue-7)', name: 'Blue' },
    { color: 'var(--oc-cyan-7)', name: 'Cyan' },
    { color: 'var(--oc-teal-7)', name: 'Teal' },
    { color: 'var(--oc-green-7)', name: 'Green' },
    { color: 'var(--oc-lime-7)', name: 'Lime' },
    { color: 'var(--oc-yellow-7)', name: 'Yellow' },
    { color: 'var(--oc-orange-7)', name: 'Orange' },
]

export const DEFAULT_ACTIVE_COLOR = 'var(--oc-grape-7)'

/** Displays custom radio group for picking color of code insight chart line. */
export const FormColorInput: React.FunctionComponent<FormColorPickerProps> = props => {
    const { className, value: propertyValue = null, name, colours = DEFAULT_COLOURS, onChange = noop } = props

    const isControlled = useRef(propertyValue !== null)
    const [internalValue, setInternalValue] = useState(DEFAULT_ACTIVE_COLOR)

    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            if (isControlled.current) {
                onChange(event)
            } else {
                setInternalValue(event.target.value)
            }
        },
        [isControlled, onChange]
    )

    const value = isControlled.current ? propertyValue : internalValue

    return (
        <div className={classnames(styles.formColorPicker, className)}>
            <div className={styles.formColorPickerColourContent}>
                {colours.map(colorInfo => (
                    <label
                        key={colorInfo.color}
                        /* eslint-disable-next-line react/forbid-dom-props */
                        style={{ color: colorInfo.color }}
                        title={colorInfo.name}
                        className={styles.formColorPickerColorBlock}
                    >
                        <input
                            type="radio"
                            name={name}
                            value={colorInfo.color}
                            checked={value === colorInfo.color}
                            className={styles.formColorPickerNativeRadioControl}
                            onChange={handleChange}
                        />

                        <span className={styles.formColorPickerRadioControl} />
                    </label>
                ))}
            </div>
        </div>
    )
}
