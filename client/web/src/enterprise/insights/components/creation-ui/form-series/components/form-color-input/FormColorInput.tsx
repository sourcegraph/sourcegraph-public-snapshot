import React, { type ChangeEventHandler, memo } from 'react'

import classNames from 'classnames'
import { startCase } from 'lodash'
import { noop } from 'rxjs'

import { Label } from '@sourcegraph/wildcard'

import { DATA_SERIES_COLORS } from '../../../../../constants'

import styles from './FormColorInput.module.scss'

interface FormColorInputProps {
    /** Name of data series color */
    name?: string
    /** Title of color input. */
    title?: string
    /** Value of data series color ()*/
    value?: string
    /** Change listener to track changes of color radio group. */
    onChange?: ChangeEventHandler<HTMLInputElement>
    /** Custom class name. */
    className?: string
}

const COLORS_KEYS = Object.keys(DATA_SERIES_COLORS) as (keyof typeof DATA_SERIES_COLORS)[]

/** Displays custom radio group for picking color of code insight chart line. */
export const FormColorInput: React.FunctionComponent<React.PropsWithChildren<FormColorInputProps>> = memo(props => {
    const { className, value = null, title, name, onChange = noop } = props

    return (
        <fieldset className={classNames('d-flex flex-column', className)}>
            <legend className={classNames('mb-3', styles.formColorPickerTitle)}>{title}</legend>

            <div>
                {COLORS_KEYS.map(key => (
                    <Label
                        key={key}
                        style={{ color: DATA_SERIES_COLORS[key] }}
                        title={startCase(key.toLocaleLowerCase())}
                        className={styles.formColorPickerColorBlock}
                    >
                        {/* eslint-disable-next-line react/forbid-elements */}
                        <input
                            type="radio"
                            name={name}
                            aria-label={key}
                            value={DATA_SERIES_COLORS[key]}
                            checked={value === DATA_SERIES_COLORS[key]}
                            className={styles.formColorPickerNativeRadioControl}
                            onChange={onChange}
                        />

                        <span className={styles.formColorPickerRadioControl} />
                    </Label>
                ))}
            </div>
        </fieldset>
    )
})

FormColorInput.displayName = 'FormColorInput'
