import { EditorState, StateField } from '@codemirror/state'
import { Tooltip, showTooltip } from '@codemirror/view'
import { render } from 'react-dom'

import { CodyRecipesWidget } from '../../../../cody/widgets/CodyRecipesWidget'

/**
 * Add a extension to CodeMirror extensions to display the Cody widget
 * when some code is selected in the editor.
 */
export function codyWidgetExtension(): StateField<readonly Tooltip[]> {
    return StateField.define<readonly Tooltip[]>({
        create: getCodyWidget,
        update(tooltips, transaction) {
            if (transaction.selection) {
                return getCodyWidget(transaction.state)
            }
            return tooltips
        },
        provide: field => showTooltip.computeN([field], state => state.field(field)),
    })
}

// @TODO: Figure out how to properly update the component instead of recreating it every time
function getCodyWidget(state: EditorState): readonly Tooltip[] {
    const { selection } = state

    if (!selection || selection.ranges.length === 0) {
        return []
    }

    const [range] = selection.ranges
    const { head, anchor } = range

    // Don't show the tooltip if the selection is collapsed
    if (head === anchor) {
        return []
    }

    const headLine = state.doc.lineAt(head)
    const anchorLine = state.doc.lineAt(anchor)

    let multiline = false
    if (headLine.number !== anchorLine.number) {
        multiline = true
    }

    // @TODO: Tweak this behavior to make it less awkward to use the tooltip
    const pos = multiline
        ? // When a multiline selection is made, the tooltip should be anchored to the start of the
          // last line. This is to avoid the tooltip from jumping around too much
          Math.max(headLine.from, anchorLine.from)
        : // Otherwise, anchor the tooltip to the end of the selection
          Math.max(head, anchor)

    return [
        {
            pos,
            above: false,
            strictSide: true,
            arrow: false,
            create: () => {
                const dom = document.createElement('div')
                dom.style.background = 'transparent'

                // @TODO: Get rid of sync render API once we figure out how to properly update the
                // component
                render(<CodyRecipesWidget />, dom)

                return { dom }
            },
        },
    ]
}
