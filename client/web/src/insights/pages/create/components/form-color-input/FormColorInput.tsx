import classnames from 'classnames';
import React, { ChangeEvent, ChangeEventHandler, ReactElement, useCallback, useRef, useState } from 'react';
import { noop } from 'rxjs';

import styles from './FormColorInput.module.scss'

interface FormColorPickerProps {
    className?: string;
    name?: string;
    value?: string;
    colours?: { color: string, name?: string, active?: boolean }[];
    onChange?: ChangeEventHandler<HTMLInputElement>
}

const DEFAULT_COLOURS = [
    { color: 'var(--oc-red-7)', name: 'red color'},
    { color: 'var(--oc-pink-7)' },
    { color: 'var(--oc-grape-7)', },
    { color: 'var(--oc-violet-7)' },
    { color: 'var(--oc-indigo-7)' },
    { color: 'var(--oc-blue-7)' },
    { color: 'var(--oc-cyan-7)' },
    { color: 'var(--oc-teal-7)' },
    { color: 'var(--oc-green-7)' },
    { color: 'var(--oc-lime-7)' },
    { color: 'var(--oc-yellow-7)' },
    { color: 'var(--oc-orange-7)' },
];

export const DEFAULT_ACTIVE_COLOR = 'var(--oc-grape-7)';

export function FormColorInput(props: FormColorPickerProps): ReactElement {
    const { className, value: propertyValue = null, name, colours = DEFAULT_COLOURS, onChange = noop } = props;

    const isControlled = useRef(propertyValue !== null);
    const [internalValue, setInternalValue] = useState(DEFAULT_ACTIVE_COLOR)

    const handleChange = useCallback(
        (event: ChangeEvent<HTMLInputElement>) => {
            if (isControlled.current) {
                onChange(event);
            } else {
                setInternalValue(event.target.value)
            }
        },
        [isControlled, onChange]
    );

    const value = isControlled.current ? propertyValue : internalValue;

    return (
        <div className={classnames(styles.formColorPicker, className)}>

            <div className={styles.formColorPickerColourContent}>

                { colours.map(colorInfo =>
                    <label
                        key={colorInfo.color}
                        /* eslint-disable-next-line react/forbid-dom-props */
                        style={{ color: colorInfo.color }}
                        title={colorInfo.name}
                        className={styles.formColorPickerColorBlock}>

                        <input
                            type="radio"
                            name={name}
                            value={colorInfo.color}
                            checked={value === colorInfo.color}
                            className={styles.formColorPickerNativeRadioControl}
                            onChange={handleChange}/>

                        <span className={styles.formColorPickerRadioControl}/>
                    </label>
                )}
            </div>

            <div>or <span className={styles.formColorPickerCustomColor}>use custom color</span></div>
        </div>
    )
}
