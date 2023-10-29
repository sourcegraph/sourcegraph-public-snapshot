import { RangeSetBuilder, type Extension } from '@codemirror/state'
import { Decoration, EditorView } from '@codemirror/view'
import { NavigateFunction } from 'react-router-dom'
import { concat, from, of } from 'rxjs'
import { timeoutWith } from 'rxjs/operators'

import { LoadingTooltip } from '../tooltips/LoadingTooltip'
import { uiPositionToOffset } from '../utils'

import { CodeIntelAPIAdapter, codeIntelAPI, getCodeIntelAPI } from './api'
import { goToDefinitionExtension } from './definition'
import { documentHighlightsExtension } from './document-highlights'
import { hoverExtension } from './hover'
import { keyboardShortcutsExtension } from './keybindings'
import { modifierClickExtension } from './modifier-click'
import { pinnedLocation } from './pin'
import { selectedToken, selectedTokenExtension } from './token-selection'
import { showTooltip, uniqueTooltips } from './tooltips'

interface CodeIntelExtensionConfig {
    api: CodeIntelAPIAdapter
    navigate: NavigateFunction
    onUnpin?: () => {}
}

const showPinnedTooltip = showTooltip.computeN([pinnedLocation], state => {
    const pin = state.facet(pinnedLocation)
    if (pin) {
        const { line = null, character = null } = pin
        if (line !== null && character !== null) {
            const offset = uiPositionToOffset(state.doc, { line, character })
            if (offset) {
                const api = getCodeIntelAPI(state)
                const range = api.findOccurrenceRangeAt(offset, state)
                if (range) {
                    const tooltip$ = from(api.getHoverTooltip(state, range))
                    return [
                        {
                            range,
                            key: 'pin',
                            source: tooltip$.pipe(
                                timeoutWith(50, concat(of(new LoadingTooltip(range.from, range.to)), tooltip$))
                            ),
                        },
                    ]
                }
            }
        }
    }
    return []
})
// Highlight tokens with tooltips
const highlightTooltipTokens = EditorView.decorations.compute([uniqueTooltips, selectedToken], state => {
    let decorations = new RangeSetBuilder<Decoration>()
    const tooltips = state.facet(uniqueTooltips)
    const selectedRange = state.field(selectedToken)
    for (const { pos, end } of tooltips) {
        // We shouldn't add/remove any decorations inside the selected token, because
        // that causes the node to be recreated and loosing focus, which breaks
        // token keyboard navigation.
        if (!(selectedRange && selectedRange.from === pos)) {
            decorations.add(pos, end, Decoration.mark({ class: `selection-highlight` }))
        }
    }

    return decorations.finish()
})

export function createCodeIntelExtension(config: CodeIntelExtensionConfig): Extension {
    return [
        codeIntelAPI.of(config.api),
        highlightTooltipTokens,
        showPinnedTooltip,
        selectedTokenExtension,
        hoverExtension,
        documentHighlightsExtension,
        modifierClickExtension,
        keyboardShortcutsExtension(config.navigate),
        goToDefinitionExtension(),
    ]
}
