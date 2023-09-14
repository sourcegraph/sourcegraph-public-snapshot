import type React from 'react'
import { useEffect, useCallback } from 'react'

import { onWindowClose } from './js-to-java-bridge'

const CAPTURE_PHASE = { capture: true }

export const GlobalKeyboardListeners: React.FunctionComponent<{}> = () => {
    const handleKeyDown = useCallback((event: KeyboardEvent) => {
        if (event.key === 'Escape' && !event.ctrlKey && !event.shiftKey && !event.altKey && !event.metaKey) {
            if (isAnyDropdownOpen()) {
                return
            }

            onWindowClose()
                .then(() => {})
                .catch(() => {})
        }
    }, [])

    useEffect(() => {
        // We're adding listeners to the capture phase to be able to examine the dropdown status before the event is
        // propagated and the dropdown is closed.
        document.addEventListener('keydown', handleKeyDown, CAPTURE_PHASE)
        return () => document.removeEventListener('keydown', handleKeyDown, CAPTURE_PHASE)
    }, [handleKeyDown])

    return null
}

// CodeMirror does not prevent bubbling of the escape keyboard event so we need to check if any dropdown is open.
export function isAnyDropdownOpen(): boolean {
    const isCodeMirrorDropdownOpen = document.querySelector('.cm-tooltip-autocomplete') !== null
    const isSearchContextDropdownOpen = document.querySelector('.jb-search-context-dropdown.show') !== null
    return isCodeMirrorDropdownOpen || isSearchContextDropdownOpen
}
