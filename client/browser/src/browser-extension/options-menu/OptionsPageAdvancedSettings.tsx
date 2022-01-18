import classNames from 'classnames'
import React from 'react'

interface OptionsPageAdvancedSettingsProps {
    optionFlags: { key: string; label: string; value: boolean }[]
    onChangeOptionFlag: (key: string, value: boolean) => void
}

export const OptionsPageAdvancedSettings: React.FunctionComponent<OptionsPageAdvancedSettingsProps> = ({
    optionFlags,
    onChangeOptionFlag,
}) => (
    <section className="mt-2">
        <ul className="p-0 m-0">
            {optionFlags.map(({ label, key, value }, index) => (
                <li className="form-check" key={key}>
                    <small>
                        <label
                            className={classNames(
                                'form-check-label cursor-pointer d-flex align-items-center font-weight-normal',
                                { 'mb-2': index !== optionFlags.length - 1 }
                            )}
                        >
                            <input
                                id={key}
                                onChange={event => onChangeOptionFlag(key, event.target.checked)}
                                className="form-check-input mb-0 mt-0"
                                type="checkbox"
                                checked={value}
                            />{' '}
                            {label}
                        </label>
                    </small>
                </li>
            ))}
        </ul>
    </section>
)
