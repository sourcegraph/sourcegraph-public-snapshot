import { useEffect, useState } from 'react'

import { HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import { EditorState, EditorStateConfig, Extension, StateEffect } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { tags } from '@lezer/highlight'

/**
 * Hook for rendering and updating a CodeMirror instance.
 */
export function useCodeMirror(
    container: HTMLDivElement | null,
    value: string,
    extensions?: EditorStateConfig['extensions']
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
    // editor. Doing this instead of setting the initial value when the state is
    // created ensures that extensions have a chance to modify the document.
    useEffect(() => {
        if (view) {
            const currentValue = view.state.sliceDoc() ?? ''

            if (currentValue !== value) {
                view.dispatch({
                    changes: { from: 0, to: currentValue.length, insert: value ?? '' },
                })
            }
        }
    }, [value, view])

    useEffect(() => {
        if (view && extensions) {
            view.dispatch({ effects: StateEffect.reconfigure.of(extensions) })
        }
        // View is not provided because this should only be triggered after the view
        // was created.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [extensions])

    return view
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
