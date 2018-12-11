import { FormattingOptions } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { SlackNotificationsConfig } from '../schema/settings.schema'
import { OpenIDConnectAuthProvider, SAMLAuthProvider, SiteConfiguration } from '../schema/site.schema'
import { parseJSON } from '../settings/configuration'
import { ConfigInsertionFunction } from '../settings/MonacoSettingsEditor'

const defaultFormattingOptions: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

const addGSuiteOIDCAuthProvider: ConfigInsertionFunction = config => {
    const value: OpenIDConnectAuthProvider = {
        type: 'openidconnect',
        issuer: 'https://accounts.google.com',
        clientID: '<see documentation: https://developers.google.com/identity/protocols/OpenIDConnect#getcredentials>',
        clientSecret: '<see same documentation as clientID>',
        requireEmailDomain: "<your company's email domain (example: mycompany.com)>",
    }
    return {
        edits: [...setProperty(config, ['auth.providers'], [value], defaultFormattingOptions)],
    }
}

const addSAMLAuthProvider: ConfigInsertionFunction = config => {
    const value: SAMLAuthProvider = {
        type: 'saml',
        identityProviderMetadataURL: '<see https://docs.sourcegraph.com/admin/auth/#saml>',
    }
    return {
        edits: [...setProperty(config, ['auth.providers'], [value], defaultFormattingOptions)],
    }
}

const addLicenseKey: ConfigInsertionFunction = config => {
    const value =
        '<input a license key generated from /site-admin/license. See https://about.sourcegraph.com/pricing for more details>'
    const edits = setProperty(config, ['licenseKey'], value, defaultFormattingOptions)
    return { edits, selectText: value }
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
    return { edits, selectText: '""', cursorOffset: 1 }
}

export interface EditorAction {
    id: string
    label: string
    run: ConfigInsertionFunction
}

export const settingsActions: EditorAction[] = [
    { id: 'sourcegraph.settings.searchScopes', label: 'Add search scope', run: addSearchScopeToSettings },
    { id: 'sourcegraph.settings.addSlackWebhook', label: 'Add Slack webhook', run: addSlackWebhook },
]

export const siteConfigActions: EditorAction[] = [
    {
        id: 'sourcegraph.site.addGSuiteOIDCAuthProvider',
        label: 'Add G Suite user auth',
        run: addGSuiteOIDCAuthProvider,
    },
    { id: 'sourcegraph.site.addSAMLAUthProvider', label: 'Add SAML user auth', run: addSAMLAuthProvider },
    { id: 'sourcegraph.site.addLicenseKey', label: 'Add license key', run: addLicenseKey },
]

export function getUpdateChannel(text: string): string {
    const channel = getProperty(text, 'update.channel')
    return channel || 'release'
}

function getProperty(text: string, property: keyof SiteConfiguration): any | null {
    try {
        const parsedConfig = parseJSON(text) as SiteConfiguration
        return parsedConfig && parsedConfig[property] !== undefined ? parsedConfig[property] : null
    } catch (err) {
        console.error(err)
        return null
    }
}
