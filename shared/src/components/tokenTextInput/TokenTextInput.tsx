import React, { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import { MultilineTextField } from '../multilineTextField/MultilineTextField'
// import { DisplayToken, Tokenizer } from 'tokenizer.js'

interface Props extends Pick<React.HTMLProps<HTMLTextAreaElement>, 'className' | 'onFocus' | 'onBlur'> {
    className?: string
    value: string
    placeholder?: string
    onChange: (newValue: string) => void
}

/**
 * A text input field that may contain a mixture of tokens and non-tokenized text.
 */
// tslint:disable: jsx-no-lambda
export const TokenTextInput: React.FunctionComponent<Props> = ({
    className = '',
    value,
    placeholder,
    onChange,
    ...props
}) => {
    const [element, setElement] = useState<HTMLDivElement | null>(null)

    // const tokenizer = useMemo(() => {
    //     if (!element) {
    //         return undefined
    //     }
    //     return new Tokenizer(element, {
    //         initialInput: tokenizeFood('banana tomato mango onio'),
    //
    //         isFocused: true,
    //         onCaretPositionChanged: () => void 0,
    //         onChange: async (inputText, caretPosition, isCaretOnSeparator) => {
    //             onChange(inputText)
    //             return asyncTokenizeFood(inputText, caretPosition)
    //         },
    //     })
    // }, [element])
    // useEffect(() => () => tokenizer && tokenizer.destroy(), [tokenizer])
    //
    // useEffect(() => {
    //     if (tokenizer && tokenizer.getInnerText() !== value) {
    //         tokenizer.updateText(value)
    //     }
    // }, [value, tokenizer])

    return (
        <MultilineTextField
            {...props}
            className={`token-text-input ${className}`}
            style={{ resize: 'none' }}
            value={value}
            placeholder={placeholder}
            onChange={e => onChange(e.currentTarget.value)}
        />
    )
}
