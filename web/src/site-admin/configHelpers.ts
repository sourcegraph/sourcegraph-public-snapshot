import { FormattingOptions } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { ConfigInsertionFunction } from '../settings/MonacoSettingsEditor'

const defaultFormattingOptions: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

const setSearchContextLines: ConfigInsertionFunction = config => {
    const DEFAULT = 3 // a reasonable value that will be clearly different from the default 1
    return { edits: setProperty(config, ['search.contextLines'], DEFAULT, defaultFormattingOptions) }
}

const addSearchScopeToSettings: ConfigInsertionFunction = config => {
    const value: { name: string; value: string } = {
        name: '<name>',
        value: '<partial query string that will be inserted when the scope is selected>',
    }
    const edits = setProperty(config, ['search.scopes', -1], value, defaultFormattingOptions)
    return { edits, selectText: '<name>' }
}

const addRepositoryGroupToSettings: ConfigInsertionFunction = config => {
    const name = prompt('Name of new repository group:')
    if (name === null) {
        return { edits: [] }
    }

    const edits = setProperty(config, ['search.repositoryGroups', name], [], defaultFormattingOptions)
    return { edits }
}

const addQuickLinkToSettings: ConfigInsertionFunction = config => {
    const value: { name: string; url: string } = {
        name: '<human-readable name>',
        url: '<URL>',
    }
    const edits = setProperty(config, ['quicklinks', -1], value, defaultFormattingOptions)
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
    { id: 'sourcegraph.settings.searchScopes', label: 'Add search scope', run: addSearchScopeToSettings },
    { id: 'sourcegraph.settings.addRepositoryGroup', label: 'Add repository group', run: addRepositoryGroupToSettings },
    { id: 'sourcegraph.settings.quickLinks', label: 'Add quick link', run: addQuickLinkToSettings },
]
