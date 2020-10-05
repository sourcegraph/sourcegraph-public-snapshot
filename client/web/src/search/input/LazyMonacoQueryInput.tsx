import React, { Suspense, useCallback, useRef } from 'react'
import { MonacoQueryInputProps } from './MonacoQueryInput'
import { lazyComponent } from '../../util/lazyComponent'
import { Toggles } from './toggles/Toggles'
import { Shortcut } from '@slimsag/react-shortcuts'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '../../keyboardShortcuts/keyboardShortcuts'

const MonacoQueryInput = lazyComponent(() => import('./MonacoQueryInput'), 'MonacoQueryInput')

/**
 * A plain query input displayed during lazy-loading of the MonacoQueryInput.
 * It has no suggestions, but still allows to type in and submit queries.
 */
export const PlainQueryInput: React.FunctionComponent<MonacoQueryInputProps> = ({
    queryState,
    autoFocus,
    onChange,
    keyboardShortcutForFocus,
    ...props
}) => {
    const onInputChange = React.useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            // cursorPosition is only used for legacy suggestions, it's OK to set it to 0 here.
            onChange({ query: event.target.value, cursorPosition: 0 })
        },
        [onChange]
    )

    const inputReference = useRef<HTMLInputElement>(null)

    const focusInputAndPositionCursorAtEnd = useCallback(() => {
        if (inputReference.current) {
            inputReference.current.focus()
            inputReference.current.setSelectionRange(
                inputReference.current.value.length,
                inputReference.current.value.length
            )
        }
    }, [])

    return (
        <div className="query-input2 d-flex">
            <input
                type="text"
                autoFocus={autoFocus}
                className="form-control code lazy-monaco-query-input--intermediate-input"
                value={queryState.query}
                onChange={onInputChange}
                spellCheck={false}
                ref={inputReference}
            />
            <div className="query-input2__toggle-container">
                <Toggles {...props} navbarSearchQuery={queryState.query} />
            </div>
            {keyboardShortcutForFocus?.keybindings.map((keybinding, index) => (
                <Shortcut key={index} {...keybinding} onMatch={focusInputAndPositionCursorAtEnd} />
            ))}
        </div>
    )
}

const USE_PLAIN_QUERY = true

/**
 * A lazily-loaded {@link MonacoQueryInput}, displaying a read-only query field as a fallback during loading.
 */
export const LazyMonacoQueryInput: React.FunctionComponent<MonacoQueryInputProps> = props =>
    USE_PLAIN_QUERY ? (
        <PlainQueryInput {...props} keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR} />
    ) : (
        <Suspense
            fallback={<PlainQueryInput {...props} keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR} />}
        >
            <MonacoQueryInput {...props} keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR} />
        </Suspense>
    )
