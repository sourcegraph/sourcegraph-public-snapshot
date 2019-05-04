import React, { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { DisplayToken, Tokenizer } from 'tokenizer.js'

interface Props {
    className?: string
    value: string
    onChange: (newValue: string) => void
}

/**
 * A text input field that may contain a mixture of tokens and non-tokenized text.
 */
export const TokenTextInput: React.FunctionComponent<Props> = ({ className = '', value, onChange }) => {
    const [element, setElement] = useState<HTMLDivElement | null>(null)

    const tokenizer = useMemo(() => {
        if (!element) {
            return undefined
        }
        return new Tokenizer(element, {
            initialInput: tokenizeFood('banana tomato mango onio'),

            isFocused: true,
            onCaretPositionChanged: () => void 0,
            onChange: async (inputText, caretPosition, isCaretOnSeparator) => {
                onChange(inputText)
                return asyncTokenizeFood(inputText, caretPosition)
            },
        })
    }, [element])
    useEffect(() => () => tokenizer && tokenizer.destroy(), [tokenizer])

    useEffect(() => {
        if (tokenizer && tokenizer.getInnerText() !== value) {
            tokenizer.updateText(value)
        }
    }, [value, tokenizer])

    return <div className={`token-text-input ${className}`} contentEditable={true} ref={setElement} />
}

// This is a dummy network call. To simulate an async
// environment.
async function asyncTokenizeFood(str: string, caretPosition = 0): Promise<DisplayToken[]> {
    return tokenizeFood(str, caretPosition)
    // const tokenized = tokenizeFood(str, caretPosition)
    // return new Promise(resolve => {
    //     setTimeout(() => resolve(tokenized), 50)
    // })
}

function tokenizeFood(str: string, caretPosition = 0): DisplayToken[] {
    const fruits = matchAll(str, /orange|mango|banana|apple/g)
    const veggies = matchAll(str, /tomato|potato|onion|ginger|lady finger/g)
    const matched = merge(fruits || [], veggies || [])
    let i = 0
    const tokens: DisplayToken[] = []
    let currentToken: DisplayToken = {
        value: '',
        isIncomplete: true,
        isExtensible: false,
        className: 'incomplete-token',
    }
    while (i < str.length) {
        const match = matched[0]
        if (match && i === match.index) {
            if (currentToken.value.trim()) {
                currentToken.value = currentToken.value.replace(/\s$/, '').replace(/^\s/, '')
                tokens.push(currentToken)
                currentToken = { ...currentToken }
            }
            currentToken.value = ''
            tokens.push({
                value: match[0],
                isIncomplete: false,
                isExtensible: false,
                className: match.type,
            })
            i += match[0].length
            matched.shift()
        } else {
            currentToken.value += str[i]
            i++
        }
    }
    if (currentToken.value.trimLeft()) {
        currentToken.value = currentToken.value.trimLeft()
        tokens.push(currentToken)
    }
    return tokens
}

function merge(fruits, veggies) {
    fruits.forEach(f => (f.type = 'fruit'))
    veggies.forEach(v => (v.type = 'veggie'))
    return fruits.concat(veggies).sort((a, b) => a.index - b.index)
}

function matchAll(str, regexp) {
    const matches = []
    str.replace(regexp, function() {
        const arr = [].slice.call(arguments, 0)
        const extras = arr.splice(-2)
        arr.index = extras[0]
        arr.input = extras[1]
        matches.push(arr)
    })
    return matches.length ? matches : null
}
