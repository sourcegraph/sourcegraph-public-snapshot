import { FormattingOptions , modify } from '@sqs/jsonc-parser'

import { ConfigInsertionFunction } from '../settings/MonacoSettingsEditor'

const formattingOptions: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

const options = { formattingOptions }

const setSearchContextLines: ConfigInsertionFunction = config => {
    const DEFAULT = 3 // a reasonable value that will be clearly different from the default 1
    return { edits: modify(config, ['search.contextLines'], DEFAULT, options) }
}

const addSearchScopeToSettings: ConfigInsertionFunction = config => {
    const value: { name: string; value: string } = {
        name: '<name>',
        value: '<partial query string that will be inserted when the scope is selected>',
    }
    const edits = modify(config, ['search.scopes', -1], value, options)
    return { edits, selectText: '<name>' }
}

const addQuickLinkToSettings: ConfigInsertionFunction = config => {
    const value: { name: string; url: string } = {
        name: '<name>',
        url: '<URL>',
    }
    const edits = modify(config, ['quicklinks', -1], value, options)
    return { edits, selectText: '<name>' }
}

export interface EditorAction {
    id: string
    label: string
    run: ConfigInsertionFunction
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
