import React, { useCallback } from 'react'
import { MultilineTextField } from '../multilineTextField/MultilineTextField'

interface Props extends Pick<React.HTMLProps<HTMLTextAreaElement>, 'className' | 'onFocus' | 'onBlur'> {
    className?: string
    value: string
    placeholder?: string
    onChange: (newValue: string) => void
}

/**
 * A text input field that may contain a mixture of tokens and non-tokenized text.
 */
export const TokenTextInput: React.FunctionComponent<Props> = ({
    className = '',
    value,
    placeholder,
    onChange: onChangeValue,
    ...props
}) => {
    const onChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => onChangeValue(e.currentTarget.value),
        [onChangeValue]
    )
    return (
        <MultilineTextField
            {...props}
            className={`token-text-input ${className}`}
            style={{ resize: 'none' }}
            value={value}
            placeholder={placeholder}
            onChange={onChange}
        />
    )
}
