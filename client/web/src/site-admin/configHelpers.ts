import { ModificationOptions, modify } from 'jsonc-parser'

import { EditorAction } from '../settings/EditorActionsGroup'
import { ConfigInsertionFunction } from '../settings/MonacoSettingsEditor'

const defaultModificationOptions: ModificationOptions = {
    formattingOptions: {
        eol: '\n',
        insertSpaces: true,
        tabSize: 2,
    },
}

const setSearchContextLines: ConfigInsertionFunction = config => {
    const DEFAULT = 3 // a reasonable value that will be clearly different from the default 1
    return { edits: modify(config, ['search.contextLines'], DEFAULT, defaultModificationOptions) }
}

const addSearchScopeToSettings: ConfigInsertionFunction = config => {
    const value: { name: string; value: string } = {
        name: '<name>',
        value: '<partial query string that will be inserted when the scope is selected>',
    }
    const edits = modify(config, ['search.scopes', -1], value, defaultModificationOptions)
    return { edits, selectText: '<name>' }
}

const addQuickLinkToSettings: ConfigInsertionFunction = config => {
    const value: { name: string; url: string } = {
        name: '<human-readable name>',
        url: '<URL>',
    }
    const edits = modify(config, ['quicklinks', -1], value, defaultModificationOptions)
    return { edits, selectText: '<name>' }
}

export const settingsActions: EditorAction[] = [
    {
        id: 'sourcegraph.settings.search.contextLines',
        label: 'Search: show # before/after lines',
        run: setSearchContextLines,
    },
    { id: 'sourcegraph.settings.searchScopes', label: 'Add search snippet', run: addSearchScopeToSettings },
    { id: 'sourcegraph.settings.quickLinks', label: 'Add quick link', run: addQuickLinkToSettings },
]
