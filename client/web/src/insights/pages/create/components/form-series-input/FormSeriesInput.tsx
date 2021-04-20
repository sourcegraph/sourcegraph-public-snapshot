import classnames from 'classnames';
import React, { ReactElement } from 'react';

import styles from './FormSeriesInput.module.scss'

interface FormColorPickerProps {
    className?: string;
    colours?: { color: string, name?: string, active?: boolean }[];
}

const DEFAULT_COLOURS = [
    { color: 'var(--oc-red-7)', name: 'red color'},
    { color: 'var(--oc-pink-7)' },
    { color: 'var(--oc-grape-7)', active: true, },
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

export function FormColorPicker(props: FormColorPickerProps): ReactElement {
    const { className, colours = DEFAULT_COLOURS } = props;

    return (
        <div className={classnames(styles.formColorPicker, className)}>

            <div className={styles.formColorPickerColourContent}>

                { colours.map(colorInfo =>
                    <div
                        key={colorInfo.color}
                        className={classnames(
                            styles.formColorPickerColorBlock,
                            { [styles.formColorPickerColorBlockActive]: colorInfo.active })
                        }
                        /* eslint-disable-next-line react/forbid-dom-props */
                        style={{ color: colorInfo.color }}
                        title={colorInfo.name}/>
                )}
            </div>

            <div>or <span className={styles.formColorPickerCustomColor}>use custom color</span></div>
        </div>
    )
}
