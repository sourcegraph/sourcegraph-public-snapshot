import React, { useCallback } from 'react'

import styles from './CodeTextArea.module.scss'

interface CodeTextAreaProps {
    placeholder?: string
    rows?: number
    onChange: (value: string) => void
    value: string
    dataTestId?: string
}

export const CodeTextArea: React.FC<CodeTextAreaProps> = ({ value, placeholder, rows = 2, onChange, dataTestId }) => {
    const handleChange: React.ChangeEventHandler<HTMLTextAreaElement> = useCallback(
        event => {
            onChange(event.target.value)
        },
        [onChange]
    )

    const lineNumbers = new Array(Math.max(value.split(/\n/).length, rows)).fill(0)

    return (
        // eslint-disable-next-line react/forbid-dom-props
        <div className={styles.container} style={{ maxHeight: `${rows * 1.6}rem` }}>
            <ul className={styles.gutter}>
                {lineNumbers.map((line, index) => (
                    <li key={index}>{index + 1}</li>
                ))}
            </ul>
            <textarea
                data-testid={dataTestId}
                rows={rows}
                value={value}
                className={styles.textarea}
                placeholder={placeholder}
                spellCheck={false}
                onChange={handleChange}
            />
        </div>
    )
}
