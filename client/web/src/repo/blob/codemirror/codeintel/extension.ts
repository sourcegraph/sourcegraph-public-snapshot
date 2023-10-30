import { type Extension } from '@codemirror/state'
import { NavigateFunction } from 'react-router-dom'
import { concat, from, of } from 'rxjs'
import { timeoutWith } from 'rxjs/operators'

import { LoadingTooltip } from '../tooltips/LoadingTooltip'
import { uiPositionToOffset } from '../utils'

import { CodeIntelAPIAdapter, codeIntelAPI, getCodeIntelAPI } from './api'
import { goToDefinitionOnClick } from './definition'
import { hoverExtension } from './hover'
import { keyboardShortcutsExtension } from './keybindings'
import { isModifierKeyHeld } from './modifier-key'
import { pinnedLocation } from './pin'
import { selectedTokenExtension } from './token-selection'
import { showTooltip } from './tooltips'

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

export function createCodeIntelExtension(config: CodeIntelExtensionConfig): Extension {
    return [
        codeIntelAPI.of(config.api),
        showPinnedTooltip,
        selectedTokenExtension,
        hoverExtension,
        isModifierKeyHeld,
        keyboardShortcutsExtension(config.navigate),
        goToDefinitionOnClick(),
    ]
}
