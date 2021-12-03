import {
    Combobox,
    ComboboxInput,
    ComboboxOption,
    ComboboxPopover,
    ComboboxOptionText,
    ComboboxList,
} from '@reach/combobox'
import classNames from 'classnames'
import { debounce } from 'lodash'
import React, { useMemo, useState, useCallback, useRef, useEffect } from 'react'

import { isModifierKeyPressed } from '../useBlockShortcuts'

import styles from './SearchNotebookFileBlockInput.module.scss'

interface SearchNotebookFileBlockInputProps {
    id?: string
    className?: string
    inputClassName?: string
    placeholder: string
    value: string
    onChange: (value: string) => void
    onFocus: (event: React.FocusEvent<HTMLInputElement>) => void
    onBlur: (event: React.FocusEvent<HTMLInputElement>) => void
    suggestions?: string[]
    suggestionsIcon?: JSX.Element
    isValid?: boolean
    isMacPlatform: boolean
    dataTestId?: string
    testTriggerSuggestions?: boolean
}

export const SearchNotebookFileBlockInput: React.FunctionComponent<SearchNotebookFileBlockInputProps> = ({
    id,
    className,
    inputClassName,
    placeholder,
    value,
    onChange,
    onFocus,
    onBlur,
    suggestions,
    suggestionsIcon,
    isValid,
    isMacPlatform,
    dataTestId,
    testTriggerSuggestions,
}) => {
    const [inputValue, setInputValue] = useState(value)
    const debouncedOnChange = useMemo(() => debounce(onChange, 300), [onChange])
    const onSelect = useCallback(
        (value: string) => {
            setInputValue(value)
            debouncedOnChange(value)
        },
        [debouncedOnChange, setInputValue]
    )

    const inputReference = useRef<HTMLInputElement>(null)
    useEffect(() => {
        if (testTriggerSuggestions) {
            inputReference.current?.focus()
        }
    }, [inputReference, testTriggerSuggestions])

    const popoverReference = useRef<HTMLDivElement>(null)
    const onKeyDown = (event: React.KeyboardEvent): void => {
        // Reach Combobox does not automatically scroll the popover when moving the selected item with a keyboard.
        // We have to do it manually by finding the currently selected option and scrolling the popover container
        // to make it visible. We are using requestAnimationFrame to prevent triggering reflows because we are
        // referencing element sizes and scroll positions.
        // Code adapted from: https://github.com/reach/reach-ui/issues/357
        window.requestAnimationFrame(() => {
            const container = popoverReference.current
            if (!container) {
                return
            }
            const element = container.querySelector<HTMLElement>('[aria-selected=true]')
            if (!element) {
                return
            }
            const top = element.offsetTop - container.scrollTop
            const bottom = container.scrollTop + container.clientHeight - (element.offsetTop + element.clientHeight)
            if (bottom < 0) {
                container.scrollTop -= bottom
            }
            if (top < 0) {
                container.scrollTop += top
            }
        })

        if (event.key === 'Escape') {
            const target = event.target as HTMLElement
            target.blur()
        } else if (
            // Allow cmd+Enter/ctrl+Enter to propagate to run the block, stop all other events
            !(event.key === 'Enter' && isModifierKeyPressed(event.metaKey, event.ctrlKey, isMacPlatform))
        ) {
            event.stopPropagation()
        }
    }

    return (
        <Combobox openOnFocus={true} onSelect={onSelect} className={className} onKeyDown={onKeyDown}>
            <ComboboxInput
                id={id}
                ref={inputReference}
                className={classNames(
                    inputClassName,
                    'form-control',
                    isValid === true && 'is-valid',
                    isValid === false && 'is-invalid'
                )}
                value={inputValue}
                placeholder={placeholder}
                onChange={event => onSelect(event.target.value)}
                onFocus={onFocus}
                onBlur={onBlur}
                data-testid={dataTestId}
            />
            {/* Only show suggestions popover for the latest input value */}
            {suggestions && value === inputValue && (
                <ComboboxPopover ref={popoverReference} className={styles.suggestionsPopover}>
                    <ComboboxList className={styles.suggestionsList}>
                        {suggestions.map(suggestion => (
                            <ComboboxOption className={styles.suggestionsOption} key={suggestion} value={suggestion}>
                                {suggestionsIcon}
                                <ComboboxOptionText />
                            </ComboboxOption>
                        ))}
                    </ComboboxList>
                </ComboboxPopover>
            )}
        </Combobox>
    )
}
