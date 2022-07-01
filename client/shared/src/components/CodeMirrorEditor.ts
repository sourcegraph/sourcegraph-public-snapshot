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
            state: EditorState.create({ extensions }),
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
