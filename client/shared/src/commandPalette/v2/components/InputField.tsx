import MagnifyIcon from 'mdi-react/MagnifyIcon'
import React, { useCallback } from 'react'

import { SourcegraphIcon } from '../../../components/SourcegraphIcon'

import styles from './InputField.module.scss'

export interface InputFieldProps {
    isNative: boolean
    value: string
    placeholder?: string
    onChange: (value: string) => void
}
export const InputField = React.forwardRef<HTMLInputElement, InputFieldProps>(
    ({ isNative, value, onChange, placeholder }, reference) => {
        const handleChange = useCallback(
            (event: React.ChangeEvent<HTMLInputElement>) => {
                onChange(event.target.value)
            },
            [onChange]
        )
        return (
            <div className={styles.inputContainer}>
                {isNative ? (
                    <MagnifyIcon className={styles.inputIcon} />
                ) : (
                    <SourcegraphIcon className={styles.inputIcon} />
                )}
                <input
                    ref={reference}
                    autoComplete="off"
                    spellCheck="false"
                    aria-autocomplete="list"
                    className={styles.input}
                    placeholder={placeholder}
                    value={value}
                    onChange={handleChange}
                    autoFocus={true}
                    type="search"
                />
            </div>
        )
    }
)
