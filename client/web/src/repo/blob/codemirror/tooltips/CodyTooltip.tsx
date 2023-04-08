import { EditorState, StateField } from '@codemirror/state'
import { Tooltip, showTooltip } from '@codemirror/view'
import ReactDOM from 'react-dom/client'

import { CodyRecipesWidget } from '@sourcegraph/cody-ui/src/widgets/CodyRecipesWidget'

/**
 * Add a extension to CodeMirror editor extensions to display the Cody widget
 * when some code is selected in the editor.
 */
export function codyWidgetExtension(): StateField<readonly Tooltip[]> {
    return StateField.define<readonly Tooltip[]>({
        create: getCodyWidget,
        update(tooltips, tr) {
            if (tr.selection) return getCodyWidget(tr.state)
            return tooltips
        },

        provide: f => showTooltip.computeN([f], state => state.field(f)),
    })
}

function getCodyWidget(state: EditorState): readonly Tooltip[] {
    const { selection } = state

    if (selection && selection.ranges.length > 0) {
        const [range] = selection.ranges
        const { head, anchor } = range

        // If something is selected, render the widget.
        if (head !== anchor) {
            return [
                {
                    pos: head,
                    above: head < anchor,
                    strictSide: true,
                    arrow: false,
                    create: () => {
                        let dom = document.createElement('div')
                        dom.style.background = 'transparent'

                        ReactDOM.createRoot(dom).render(<CodyRecipesWidget />)
                        return { dom }
                    },
                },
            ]
        }
    }
    return []
}
