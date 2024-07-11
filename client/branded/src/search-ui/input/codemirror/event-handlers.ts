// Instead of deriving extensions directly form props, these event handlers are
// configured via a field. This means that their values can be updated via
// transactions instead of having to reconfigure the whole editor. This is
// especially useful if the event handlers are not stable across re-renders.
// Instead of creating a separate field for every handler, all handlers are set
// via a single field to keep complexity manageable.

import { Facet, Prec } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'

import { createUpdateableField } from '@sourcegraph/shared/src/components/codemirror/utils'

import { multiLineEnabled } from './multiline'

export interface QueryInputEventHandlers {
    /**
     * Called when Enter or Mod-Enter are pressed. For convenience, if multi-line mode is
     * _enabled_ it will only be called on Mod-Enter.
     */
    onEnter?: (view: EditorView) => boolean
    onChange?: (value: string) => void
    onFocus?: (view: EditorView) => void
    onBlur?: (view: EditorView) => void
}

const eventHandlers = Facet.define<QueryInputEventHandlers, QueryInputEventHandlers>({
    combine(value) {
        return value[0] ?? {}
    },
    enables(self) {
        return [
            Prec.high(
                keymap.of([
                    {
                        key: 'Enter',
                        run: view =>
                            !view.state.facet(multiLineEnabled) && (view.state.facet(self).onEnter?.(view) ?? false),
                    },
                    {
                        key: 'Mod-Enter',
                        run: view => view.state.facet(self).onEnter?.(view) ?? false,
                    },
                ])
            ),
            EditorView.updateListener.of(update => {
                const { state, view } = update
                const { onChange, onFocus, onBlur } = state.facet(self)

                if (update.docChanged) {
                    onChange?.(state.sliceDoc())
                }

                // The focus and blur event handlers are implemented via state update handlers
                // because it appears that binding them as DOM event handlers triggers them at
                // the moment they are bound if the editor is already in that state ((not)
                // focused). See https://github.com/sourcegraph/sourcegraph/issues/37721#issuecomment-1166300433
                if (update.focusChanged) {
                    if (view.hasFocus) {
                        onFocus?.(view)
                    } else {
                        onBlur?.(view)
                    }
                }
            }),
        ]
    },
})

const [searchInputEventHandlers, setSearchInputEventHandlers] = createUpdateableField<QueryInputEventHandlers>(
    {},
    field => eventHandlers.from(field)
)

export { searchInputEventHandlers, setSearchInputEventHandlers }
