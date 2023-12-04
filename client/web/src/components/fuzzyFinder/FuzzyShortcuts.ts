import type { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'

import { useFuzzyFinderFeatureFlags } from './FuzzyFinderFeatureFlag'
import type { FuzzyTabKey } from './FuzzyTabs'

interface FuzzyShortcut {
    name: FuzzyTabKey
    isEnabled: boolean
    shortcut: KeyboardShortcut | undefined
}

export function useFuzzyShortcuts(): FuzzyShortcut[] {
    const { fuzzyFinderActions, fuzzyFinderRepositories, fuzzyFinderSymbols } = useFuzzyFinderFeatureFlags()
    const fuzzyFinderShortcut = useKeyboardShortcut('fuzzyFinder')
    const fuzzyFinderActionsShortcut = useKeyboardShortcut('fuzzyFinderActions')
    const fuzzyFinderReposShortcut = useKeyboardShortcut('fuzzyFinderRepos')
    const fuzzyFinderFilesShortcut = useKeyboardShortcut('fuzzyFinderFiles')
    const fuzzyFinderSymbolsShortcut = useKeyboardShortcut('fuzzyFinderSymbols')
    return [
        {
            name: 'all',
            isEnabled: Boolean(fuzzyFinderShortcut),
            shortcut: fuzzyFinderShortcut,
        },
        {
            name: 'actions',
            isEnabled: (fuzzyFinderActions && !!fuzzyFinderActionsShortcut) || false,
            shortcut: fuzzyFinderActionsShortcut,
        },
        {
            name: 'repos',
            isEnabled: (fuzzyFinderRepositories && !!fuzzyFinderReposShortcut) || false,
            shortcut: fuzzyFinderReposShortcut,
        },
        {
            name: 'symbols',
            isEnabled: (fuzzyFinderSymbols && !!fuzzyFinderSymbolsShortcut) || false,
            shortcut: fuzzyFinderSymbolsShortcut,
        },
        {
            name: 'files',
            isEnabled: !!fuzzyFinderFilesShortcut,
            shortcut: fuzzyFinderFilesShortcut,
        },
    ]
}
