import { EditorSelection, EditorState, Extension, StateEffect, StateField } from '@codemirror/state'
import { EditorView, PluginValue, Tooltip, ViewPlugin, showTooltip } from '@codemirror/view'
import ReactDOM from 'react-dom/client'

import { CodyRecipesWidget } from '../../../../cody/widgets/CodyRecipesWidget'

export const codyTooltip = StateField.define<Tooltip | null>({
    create() {
        return null
    },
    update(value, transaction) {
        if (isSelectionCollapsed(transaction.newSelection)) {
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

    if (isSelectionCollapsed(selection)) {
        return null
    }

    const [range] = selection.ranges
    const { head, anchor } = range

    const headLine = state.doc.lineAt(head)
    const anchorLine = state.doc.lineAt(anchor)

    let multiline = false
    if (headLine.number !== anchorLine.number) {
        multiline = true
    }

    // Find a position that is always the left most position of the selection bounding box
    const pos = multiline
        ? // When a multiline selection is made, the tooltip should be anchored to the start of the
          // last line. This is to avoid the tooltip from jumping around too much
          Math.max(headLine.from, anchorLine.from)
        : // Otherwise, anchor the tooltip to the start of the selection
          Math.min(head, anchor)

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

function isSelectionCollapsed(selection: EditorSelection): boolean {
    if (!selection || selection.ranges.length === 0) {
        return true
    }

    const [range] = selection.ranges
    const { head, anchor } = range

    // Don't show the tooltip if the selection is collapsed
    if (head === anchor) {
        return true
    }

    return false
}
