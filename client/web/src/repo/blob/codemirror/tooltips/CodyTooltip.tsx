import { EditorState, Extension, StateEffect, StateField } from '@codemirror/state'
import { EditorView, PluginValue, Tooltip, ViewPlugin, showTooltip } from '@codemirror/view'
import ReactDOM from 'react-dom/client'

import { CodyRecipesWidget } from '../../../../cody/widgets/CodyRecipesWidget'

export const codyTooltip = StateField.define<Tooltip | null>({
    create() {
        return null
    },
    update(value, transaction) {
        if (transaction.newSelection.main.empty) {
            return null
        }

        for (const effect of transaction.effects) {
            if (effect.is(setCodyTooltip)) {
                return effect.value
            }
        }
        return value
    },
    provide(field) {
        return showTooltip.compute([field], state => state.field(field))
    },
})

const setCodyTooltip = StateEffect.define<Tooltip | null>()

/**
 * Add a extension to CodeMirror extensions to display the Cody widget
 * when some code is selected in the editor.
 */
export function codyWidgetExtension(): Extension {
    return [codyTooltip, selectionChangedPlugin]
}

// We use this custom plugin over EditorView.domEventHandlers() because mouse selections can start
// inside CodeMirror but the mouseup event can handle _outside_ of the CodeMirror element. These
// events still change the selection inside CodeMirror but won't be fired when using the built-in
// dom handler.
const selectionChangedPlugin = ViewPlugin.fromClass(
    class implements PluginValue {
        constructor(public view: EditorView) {
            document.body.addEventListener('mouseup', this.onPotentialSelectionChanged)
            document.body.addEventListener('keyup', this.onPotentialSelectionChanged)
        }
        public destroy(): void {
            document.body.removeEventListener('mouseup', this.onPotentialSelectionChanged)
            document.body.removeEventListener('keyup', this.onPotentialSelectionChanged)
        }
        public onPotentialSelectionChanged = (): void => {
            this.view.dispatch({ effects: [setCodyTooltip.of(computeCodyWidget(this.view.state))] })
        }
    }
)

function computeCodyWidget(state: EditorState): Tooltip | null {
    const { selection } = state

    if (selection.main.empty) {
        return null
    }

    // Find a position that is always the left most position of the selection bounding box
    const lineFrom = state.doc.lineAt(selection.main.from)
    const lineTo = state.doc.lineAt(selection.main.to)
    const isMultiline = lineFrom.number !== lineTo.number
    // When a multiline selection is made, the tooltip should be anchored to the start of the last
    // line. This is to avoid the tooltip from jumping around too much. Otherwise, anchor the
    // tooltip to the start of the selection
    const pos = isMultiline ? lineTo.from : selection.main.from

    return {
        pos,
        above: false,
        strictSide: true,
        arrow: false,
        create: () => {
            const dom = document.createElement('div')
            dom.style.background = 'transparent'
            ReactDOM.createRoot(dom).render(<CodyRecipesWidget />)
            return { dom }
        },
    }
}
