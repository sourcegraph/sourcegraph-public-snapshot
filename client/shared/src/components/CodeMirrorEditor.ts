import { useEffect, useState } from 'react'

import { EditorState, EditorStateConfig, StateEffect } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

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
            state: EditorState.create({ doc: value ?? '', extensions }),
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

    // Update editor value if necessary
    useEffect(() => {
        const currentValue = view?.state.doc.toString() ?? ''
        if (view && currentValue !== value) {
            view.dispatch({
                changes: { from: 0, to: currentValue.length, insert: value ?? '' },
            })
        }
        // View is not provided because this should only be triggered after the view
        // was created.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [value])

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
