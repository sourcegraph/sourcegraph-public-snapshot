import React from 'react'

import classNames from 'classnames'

import { Checkbox } from '@sourcegraph/wildcard'

import styles from './OptionsPageAdvancedSettings.module.scss'

interface OptionsPageAdvancedSettingsProps {
    optionFlags: { key: string; label: string; value: boolean }[]
    onChangeOptionFlag: (key: string, value: boolean) => void
}

export const OptionsPageAdvancedSettings: React.FunctionComponent<
    React.PropsWithChildren<OptionsPageAdvancedSettingsProps>
> = ({ optionFlags, onChangeOptionFlag }) => (
    <section className="mt-2">
        <ul className={classNames(styles.list, 'p-0 m-0')}>
            {optionFlags.map(({ label, key, value }, index) => (
                <li key={key}>
                    <small>
                        <Checkbox
                            id={key}
                            onChange={event => onChangeOptionFlag(key, event.target.checked)}
                            label={label}
                            checked={value}
                            className="mb-0 mt-0"
                            wrapperClassName={classNames(
                                'cursor-pointer d-flex align-items-center font-weight-normal',
                                { 'mb-2': index !== optionFlags.length - 1 }
                            )}
                        />
                    </small>
                </li>
            ))}
        </ul>
    </section>
)
