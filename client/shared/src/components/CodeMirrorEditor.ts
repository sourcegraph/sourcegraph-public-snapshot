/* eslint-disable jsdoc/check-indentation */
import { useEffect, useMemo, useState } from 'react'

import { HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import {
    ChangeSpec,
    Compartment,
    EditorState,
    EditorStateConfig,
    Extension,
    StateEffect,
    StateEffectType,
    StateField,
} from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { tags } from '@lezer/highlight'

if (process.env.INTEGRATION_TESTS) {
    // Expose findFromDOM on the global object to be able to get the real input
    // value in integration tests
    // Typecast "as any" is used to avoid TypeScript complaining about window
    // not having this property. We decided that it's fine to use this in a test
    // context
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access,@typescript-eslint/no-explicit-any
    ;(window as any).CodeMirrorFindFromDOM = (element: HTMLElement): ReturnType<typeof EditorView['findFromDOM']> =>
        EditorView.findFromDOM(element)
}

/**
 * Hook for rendering and updating a CodeMirror instance.
 */
export function useCodeMirror(
    container: HTMLDivElement | null,
    value: string,
    extensions?: EditorStateConfig['extensions'],
    options?: {
        /**
         * When 'value' changes, trigger a transaction to update it. This is `true` by default.
         * However, if other parts of the editor state should be changed when the value changes,
         * you can set this to `false` and use the `replaceValue` function to update the value
         * in a custom transaction.
         */
        updateValueOnChange?: boolean

        /**
         * When 'extension' changes, trigger a transaction to update it. This is `true` by default.
         * Set this to  `false` to have more control over how to update the editor. This is
         * useful for example when the caller wants to update the editor with `setState`.
         */
        updateOnExtensionChange?: boolean
    }
): EditorView | undefined {
    const [view, setView] = useState<EditorView>()

    useEffect(() => {
        if (!container) {
            return
        }

        const view = new EditorView({
            state: EditorState.create({ doc: value, extensions }),
            parent: container,
        })
        setView(view)
        return () => {
            setView(undefined)
            view.destroy()
        }
        // Extensions and value are updated via transactions below
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [container])

    // Update editor value if necessary. This also sets the intial value of the
    // editor.
    useEffect(() => {
        if (view && options?.updateValueOnChange !== false) {
            const changes = replaceValue(view, value ?? '')

            if (changes) {
                view.dispatch({ changes })
            }
        }
        // View is not provided because this should only be triggered after the view
        // was created.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [value, options?.updateValueOnChange])

    useEffect(() => {
        if (view && extensions && options?.updateOnExtensionChange !== false) {
            view.dispatch({ effects: StateEffect.reconfigure.of(extensions) })
        }
        // View is not provided because this should only be triggered after the view
        // was created.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [extensions, options?.updateOnExtensionChange])

    return view
}

/**
 * Create a {@link ChangeSpec} for replacing the current editor value. Returns `undefined` if the
 * new value is the same as the current value.
 */
export function replaceValue(view: EditorView, newValue: string): ChangeSpec | undefined {
    const currentValue = view.state.sliceDoc() ?? ''
    if (currentValue === newValue) {
        return undefined
    }

    return { from: 0, to: currentValue.length, insert: newValue }
}

/**
 * Helper hook for extensions that depend on on some input props.
 * With this hook the extension is isolated in a compartment so it can be
 * updated without reconfiguring the whole editor.
 *
 * Use `useMemo` to compute the extension from some input and `useEffect` to
 * update it:
 *
 * const extension = useMemo(() => EditorView.darkTheme(isLightTheme === false), [isLightTheme])
 * const [compartment, updateCompartment] = useCompartment(extension)
 * const editor = useCodeMirror(..., ..., compartment)
 *
 * useEffect(() => {
 *   if (editor) {
 *     updateCompartment(extension)
 *  }
 *}, [editor, extension])
 *
 * @param initialExtension - the extension to use when creating the editor
 *
 * @returns A compartmentalized extension and a function to update the
 * compartment
 */
export function useCompartment(
    initialExtension: Extension
): [Extension, (editor: EditorView, extension: Extension) => void] {
    return useMemo(() => {
        const compartment = new Compartment()
        return [
            compartment.of(initialExtension),
            (editor, extension: Extension) => {
                // This check avoids an unnecessary update when the editor is
                // first created
                if (initialExtension !== extension) {
                    editor.dispatch({ effects: compartment.reconfigure(extension) })
                }
            },
        ]
        // initialExtension is intentionally ignored in subsequent renders
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])
}

/**
 * A helper function for creating an extension that operates on the value which
 * can be updated via an effect.
 * This is useful in React components where the extension depends on the value
 * of a prop  but that prop is unstable, and especially useful for callbacks.
 * Instead of reconfiguring the editor whenever the value changes (which is
 * apparently not cheap), the extension can be updated via the returned update
 * function or effect.
 *
 * Example:
 *
 * const {onChange} = props;
 * const [onChangeField, setOnChange] = useMemo(() => createUpdateableField(...), [])
 * ...
 * useEffect(() => {
 *   if (editor) {
 *     setOnchange(editor, onChange)
 *   }
 * }, [editor, onChange])
 */
export function createUpdateableField<T>(
    defaultValue: T,
    provider?: (field: StateField<T>) => Extension
): [StateField<T>, (editor: EditorView, newValue: T) => void, StateEffectType<T>] {
    const fieldEffect = StateEffect.define<T>()
    const field = StateField.define<T>({
        create() {
            return defaultValue
        },
        update(value, transaction) {
            const effect = transaction.effects.find((effect): effect is StateEffect<typeof defaultValue> =>
                effect.is(fieldEffect)
            )
            return effect ? effect.value : value
        },
        provide: provider,
    })

    return [field, (editor, newValue) => editor.dispatch({ effects: [fieldEffect.of(newValue)] }), fieldEffect]
}

/**
 * Sets the height and/or max height of the editor, with corresponding overflow
 * behavior. The values can be any valid CSS unit.
 * Taken from https://codemirror.net/examples/styling/#overflow-and-scrolling
 */
export function editorHeight({
    height = null,
    maxHeight = null,
}: {
    height?: string | null
    maxHeight?: string | null
}): Extension {
    return EditorView.theme({
        '&': {
            height,
            maxHeight,
        },
        '.cm-scroller': {
            overflow: 'auto',
        },
    })
}

/**
 * Default editor theme (background color, text color, gutter color, etc)
 */
export const defaultEditorTheme = EditorView.theme({
    '&.cm-focused': {
        // CodeMirror shows a focus ring by default. Since we handle it
        // differently, disable it here.
        outline: 'none',
    },
    '.cm-content': {
        backgroundColor: 'var(--color-bg-1)',
        color: 'var(--search-query-text-color)',
        caretColor: 'var(--search-query-text-color)',
        fontFamily: 'var(--code-font-family)',
        fontSize: 'var(--code-font-size)',
    },
    '.cm-gutters': {
        backgroundColor: 'var(--color-bg-2)',
        borderColor: 'var(--border-color)',
        color: 'var(--text-muted)',
    },
    '.cm-foldPlaceholder': {
        backgroundColor: 'var(--color-bg-3)',
    },
})

/**
 * Default CodeMirror syntax highlight theme that maps highlighting tags to our
 * default CSS classes from highlight.css
 * See https://lezer.codemirror.net/docs/ref/#highlight
 */
export const defaultSyntaxHighlighting: Extension = syntaxHighlighting(
    HighlightStyle.define([
        { tag: tags.comment, class: 'hljs-comment' },
        { tag: tags.variableName, class: 'hljs-variable' },
        { tag: tags.name, class: 'hljs-name' },
        { tag: tags.keyword, class: 'hljs-keyword' },
        { tag: tags.quote, class: 'hljs-quote' },
        { tag: tags.tagName, class: 'hljs-selector-tag' },
        { tag: tags.tagName, class: 'hljs-tag' },
        { tag: tags.string, class: 'hljs-string' },
        { tag: tags.heading, class: 'hljs-title' },
        { tag: [tags.attributeName, tags.propertyName], class: 'hljs-attr' },
        { tag: tags.literal, class: 'hljs-literal' },
        { tag: tags.typeName, class: 'hljs-type' },
        { tag: tags.number, class: 'hljs-number' },
        { tag: tags.link, class: 'hljs-link' },
        { tag: tags.url, class: 'hljs-link' },
        { tag: tags.emphasis, class: 'hljs-italic' },
        { tag: tags.strong, class: 'hljs-strong' },
    ])
)

/**
 * JSON specific syntax highlighting. Extends {@link defaultSyntaxHighlighting}
 * See https://github.com/lezer-parser/json/blob/d9c5a140900134bc511bd73db3e1d81ca19a5d4f/src/highlight.js
 * for which highlighting tags are used by the default JSON parser.
 */
export const jsonHighlighting: Extension = [
    syntaxHighlighting(HighlightStyle.define([{ tag: [tags.bool, tags.null], class: 'hljs-attr' }])),
    defaultSyntaxHighlighting,
]
