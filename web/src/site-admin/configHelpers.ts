import { FormattingOptions } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { SlackNotificationsConfig } from '../schema/settings.schema'
import { ConfigInsertionFunction } from '../settings/MonacoSettingsEditor'

const defaultFormattingOptions: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

const setSearchContextLines: ConfigInsertionFunction = config => {
    const DEFAULT = 3 // a reasonable value that will be clearly different from the default 1
    return { edits: setProperty(config, ['search.contextLines'], DEFAULT, defaultFormattingOptions), selectText: `${DEFAULT}` }
}

const addSearchScopeToSettings: ConfigInsertionFunction = config => {
    const value: { name: string; value: string } = {
        name: '<name>',
        value: '<partial query string that will be inserted when the scope is selected>',
    }
    const edits = setProperty(config, ['search.scopes', -1], value, defaultFormattingOptions)
    return { edits, selectText: '<name>' }
}

const addSlackWebhook: ConfigInsertionFunction = config => {
    const value: SlackNotificationsConfig = {
        webhookURL: 'get webhook URL at https://YOUR-WORKSPACE-NAME.slack.com/apps/new/A0F7XDUAZ-incoming-webhooks',
    }
    const edits = setProperty(config, ['notifications.slack'], value, defaultFormattingOptions)
    return { edits, selectText: 'YOUR-WORKSPACE-NAME' }
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
    { id: 'sourcegraph.settings.addSlackWebhook', label: 'Add Slack webhook', run: addSlackWebhook },
]

export const siteConfigActions: EditorAction[] = []
