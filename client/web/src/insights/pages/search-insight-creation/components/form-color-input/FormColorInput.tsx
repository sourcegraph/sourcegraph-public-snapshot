import classnames from 'classnames'
import openColor from 'open-color'
import React, { ChangeEvent, ChangeEventHandler, useCallback, useRef, useState } from 'react'
import { noop } from 'rxjs'

import styles from './FormColorInput.module.scss'

interface FormColorPickerProps {
    /** Name of data series color */
    name?: string
    /** Title of color input. */
    title?: string
    /** Value of data series color ()*/
    value?: string
    /** Different values of preset color. */
    colours?: { color: string; name?: string }[]
    /** Change listener to track changes of color radio group. */
    onChange?: ChangeEventHandler<HTMLInputElement>
    /** Custom class name. */
    className?: string
}

const DEFAULT_COLOURS = Object.keys(openColor)
    .filter(name => name !== 'white' && name !== 'black')
    .map(name => ({ name: startCase(name), color: `var(--oc-${name}-7)` }))

export const DEFAULT_ACTIVE_COLOR = openColor.grape[7]

/** Displays custom radio group for picking color of code insight chart line. */
export const FormColorInput: React.FunctionComponent<FormColorPickerProps> = props => {
    const { className, value: propertyValue = null, title, name, colours = DEFAULT_COLOURS, onChange = noop } = props

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
        <fieldset className={classnames('d-flex flex-column', className)}>
            <legend className={classnames('mb-3', styles.formColorPickerTitle)}>{title}</legend>

            <div>
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
        </fieldset>
    )
}
