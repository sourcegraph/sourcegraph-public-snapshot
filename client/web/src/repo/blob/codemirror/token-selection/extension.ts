import { type Extension } from '@codemirror/state'
import { NavigateFunction } from 'react-router-dom'

import { uiPositionToOffset } from '../utils'

import { CodeIntelAPIAdapter, codeIntelAPI, getCodeIntelAPI } from './api'
import { goToDefinitionExtension } from './definition'
import { documentHighlightsExtension } from './document-highlights'
import { hoverExtension } from './hover'
import { keyboardShortcutsExtension } from './keybindings'
import { modifierClickExtension } from './modifier-click'
import { pinnedLocation } from './pin'
import { selectedTokenExtension } from './token-selection'
import { showTooltips, tooltipsExtension } from './tooltips'

interface CodeIntelExtensionConfig {
    api: CodeIntelAPIAdapter
    navigate: NavigateFunction
    onUnpin?: () => {}
}

const showPinnedTooltip = showTooltips.computeN([pinnedLocation], state => {
    const pin = state.facet(pinnedLocation)
    if (pin) {
        const { line = null, character = null } = pin
        if (line !== null && character !== null) {
            const offset = uiPositionToOffset(state.doc, { line, character })
            if (offset) {
                const range = getCodeIntelAPI(state).findOccurrenceRangeAt(offset, state)
                if (range) {
                    return [{ range, key: 'pin' }]
                }
            }
        }
    }
    return []
})

const tooltips = tooltipsExtension({ tooltipPriority: ['pin', 'focus', 'hover'] })

export function createCodeIntelExtension(config: CodeIntelExtensionConfig): Extension {
    return [
        codeIntelAPI.of(config.api),
        selectedTokenExtension,
        tooltips,
        hoverExtension,
        documentHighlightsExtension,
        modifierClickExtension,
        keyboardShortcutsExtension(config.navigate),
        goToDefinitionExtension(),
        showPinnedTooltip,
    ]
}
