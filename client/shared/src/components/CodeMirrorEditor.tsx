import React, {
    forwardRef,
    type MutableRefObject,
    type RefObject,
    useEffect,
    useImperativeHandle,
    useMemo,
    useRef,
} from 'react'

import { HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import {
    type ChangeSpec,
    Compartment,
    EditorState,
    type EditorStateConfig,
    type Extension,
    StateEffect,
    type StateEffectType,
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

const defaultTheme = EditorView.baseTheme({
    // Overwrites the default cursor color, which has too low contrast in dark mode
    '.theme-dark & .cm-cursor': {
        borderLeftColor: 'var(--grey-07)',
    },
})

/**
 * Hook for rendering and updating a CodeMirror instance.
 */
export function useCodeMirror(
    editorRef: MutableRefObject<EditorView | null>,
    containerRef: RefObject<HTMLDivElement | null>,
    value: string,
    extensions?: EditorStateConfig['extensions']
): void {
    const allExtensions = useMemo(() => [defaultTheme, extensions ?? []], [extensions])

    // The order of effects is important here:
    //
    // - If the editor hasn't been created yet (editorRef.current is null) it should be
    //   fully instantiated with value and extensions. The value/extension update effects
    //   should not have any ... effect.
    // - When the hook runs on subsequent renders the value and extensions get update if
    //   the respective values changed.
    //
    // We achieve this by putting the update effects before the creation effect.

    // Update editor value if necessary
    useEffect(() => {
        if (editorRef.current) {
            const changes = replaceValue(editorRef.current, value ?? '')

            if (changes) {
                editorRef.current.dispatch({ changes })
            }
        }
    }, [editorRef, value])

    // Reconfigure/update extensions if necessary
    useEffect(() => {
        if (editorRef.current) {
            editorRef.current.dispatch({ effects: StateEffect.reconfigure.of(allExtensions) })
        }
    }, [editorRef, allExtensions])

    // Create editor if necessary
    useEffect(() => {
        if (!editorRef.current && containerRef.current) {
            editorRef.current = new EditorView({
                state: EditorState.create({ doc: value, extensions: allExtensions }),
                parent: containerRef.current,
            })
        }
    }, [editorRef, containerRef, value, allExtensions])

    // Clean up editor on unmount
    useEffect(
        () => () => {
            editorRef.current?.destroy()
            editorRef.current = null
        },
        [editorRef]
    )
}

/**
 * Generic interface for interaction with the editor.
 */
export interface Editor {
    focus(): void
    blur(): void
}

/**
 * Helper function for creating an {@link Editor} from a {@link EditorView}.
 * When calling `.focus` the input will only be focused if it doesn't have
 * focus and it will move the cursor to the end of the input.
 */
export function viewToEditor(viewRef: RefObject<EditorView>): Editor {
    return {
        focus() {
            const view = viewRef.current
            if (view && !view.hasFocus) {
                view.focus()
                view.dispatch({
                    selection: { anchor: view.state.doc.length },
                    scrollIntoView: true,
                })
            }
        },
        blur() {
            viewRef.current?.contentDOM.blur()
        },
    }
}

/**
 * Simple React component around useCodeMirror. Use this if you have a simple setup and/or need
 * to render an editor conditionally.
 */
export const CodeMirrorEditor = React.memo(
    forwardRef<Editor, { value: string; extensions?: Extension; className?: string }>(
        ({ value, extensions, className }, ref) => {
            const containerRef = useRef<HTMLDivElement | null>(null)
            const editorRef = useRef<EditorView | null>(null)
            useCodeMirror(editorRef, containerRef, value, extensions)

            useImperativeHandle(ref, () => viewToEditor(editorRef), [])

            return <div ref={containerRef} className={className} />
        }
    )
)

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
 * Use `useMemo` to compute the extension from some input:
 *
 * const extension = useCompartment(
 * editorRef,
 * useMemo(() => EditorView.darkTheme(isLightTheme === false), [isLightTheme])
 * )
 * const editor = useCodeMirror(..., ..., extension)
 * @param editorRef - Ref object to the editor instance
 * @param extension - the dynamic extension(s) to add to the editor
 * @returns a compartmentalized extension
 */
export function useCompartment(
    editorRef: RefObject<EditorView>,
    extension: Extension,
    extender?: (view: EditorView) => StateEffect<unknown>[]
): Extension {
    const compartment = useMemo(() => new Compartment(), [])
    // We only want to trigger CodeMirror transactions when the component updates,
    // not on the first render.
    const shouldUpdate = useRef(false)

    useEffect(() => {
        const view = editorRef.current
        if (view && compartment.get(view.state) !== extension) {
            view.dispatch({ effects: [compartment.reconfigure(extension), ...(extender?.(view) ?? [])] })
        }
    }, [shouldUpdate, compartment, editorRef, extension, extender])

    // The compartment is initialized only in the first render.
    // In subsequent renders we dispatch effects to update the compartment (
    // see below).
    // eslint-disable-next-line react-hooks/exhaustive-deps
    const initialExtension = useMemo(() => compartment.of(extension), [compartment])

    return initialExtension
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
