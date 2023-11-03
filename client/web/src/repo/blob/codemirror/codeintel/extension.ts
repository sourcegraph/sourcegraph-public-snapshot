import { type Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { NavigateFunction } from 'react-router-dom'

import { uiPositionToOffset } from '../utils'

import { CodeIntelAPIAdapter, CodeIntelAPIConfig, codeIntelAPI, findOccurrenceRangeAt } from './api'
import { goToDefinitionOnClick } from './definition'
import { hoverExtension } from './hover'
import { keyboardShortcutsExtension } from './keybindings'
import { isModifierKeyHeld } from './modifier-key'
import { PinConfig, pinConfig, pinnedLocation, pinnedRange } from './pin'
import { selectedTokenExtension } from './token-selection'

interface CodeIntelExtensionConfig {
    api: CodeIntelAPIConfig
    pin: PinConfig
    navigate: NavigateFunction
}

/**
 * Registers a tooltip for the pinned location, if any. This compouted here
 * and not when providing {@link pinnedLocation} because the code intel API
 * might not be ready at that moment.
 */
const pinnedLocationToRange = pinnedRange.compute([pinnedLocation], state => {
    const pin = state.facet(pinnedLocation)
    if (pin) {
        const { line = null, character = null } = pin
        if (line !== null && character !== null) {
            const offset = uiPositionToOffset(state.doc, { line, character })
            if (offset) {
                return findOccurrenceRangeAt(state, offset)
            }
        }
    }
    return null
})

/**
 * Additional styles for loading/temporary tooltips.
 */
const tooltipStyles = EditorView.theme({
    // Tooltip styles is a combination of the default wildcard PopoverContent component (https://github.com/sourcegraph/sourcegraph/blob/5de30f6fa1c59d66341e4dfc0c374cab0ad17bff/client/wildcard/src/components/Popover/components/popover-content/PopoverContent.module.scss#L1-L10)
    // and the floating tooltip-like storybook usage example (https://github.com/sourcegraph/sourcegraph/blob/5de30f6fa1c59d66341e4dfc0c374cab0ad17bff/client/wildcard/src/components/Popover/story/Popover.story.module.scss#L54-L62)
    // ignoring the min/max width rules.
    '.cm-tooltip.tmp-tooltip': {
        fontSize: '0.875rem',
        backgroundClip: 'padding-box',
        backgroundColor: 'var(--dropdown-bg)',
        border: '1px solid var(--dropdown-border-color)',
        borderRadius: 'var(--popover-border-radius)',
        color: 'var(--body-color)',
        boxShadow: 'var(--dropdown-shadow)',
        padding: '0.5rem',
    },

    '.cm-tooltip.cm-tooltip-above.tmp-tooltip .cm-tooltip-arrow:before': {
        borderTopColor: 'var(--dropdown-border-color)',
    },
    '.cm-tooltip.cm-tooltip-above.tmp-tooltip .cm-tooltip-arrow:after': {
        borderTopColor: 'var(--dropdown-bg)',
    },
    '.cm-tooltip.sg-code-intel-hovercard': {
        border: 'unset',
    },
})

/**
 * Adds various code intel features:
 * - token navigation
 * - hover tooltips
 * - selected token tooltips
 * - pinned tooltips
 * - document highlights
 * - "go to definition"
 */
export function createCodeIntelExtension(config: CodeIntelExtensionConfig): Extension {
    return [
        codeIntelAPI.of(new CodeIntelAPIAdapter(config.api)),
        pinConfig.of(config.pin),
        goToDefinitionOnClick(),
        pinnedLocationToRange,
        selectedTokenExtension,
        hoverExtension,
        isModifierKeyHeld,
        keyboardShortcutsExtension(config.navigate),
        tooltipStyles,
    ]
}
