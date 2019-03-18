import { FormattingOptions } from '@sqs/jsonc-parser'
import { removeProperty, setProperty } from '@sqs/jsonc-parser/lib/edit'
import { SlackNotificationsConfig } from '../schema/settings.schema'
import {
    AwsCodeCommitConnection,
    BitbucketServerConnection,
    GitHubConnection,
    GitLabConnection,
    OpenIdConnectAuthProvider,
    Repository,
    SamlAuthProvider,
    SiteConfiguration,
} from '../schema/site.schema'
import { parseJSON } from '../settings/configuration'
import { ConfigInsertionFunction } from '../settings/MonacoSettingsEditor'

const defaultFormattingOptions: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

const addGitHubDotCom: ConfigInsertionFunction = config => {
    const tokenPlaceholder = '<personal access token with repo scope (https://github.com/settings/tokens/new)>'
    const value: GitHubConnection = {
        token: tokenPlaceholder,
        url: 'https://github.com',
    }
    const edits = setProperty(config, ['github', -1], value, defaultFormattingOptions)
    return { edits, selectText: tokenPlaceholder }
}

const addGitHubEnterprise: ConfigInsertionFunction = config => {
    const tokenPlaceholder = '<personal access token with repo scope>'
    const value: GitHubConnection = {
        token: tokenPlaceholder,
        url: 'https://github-enterprise-hostname.example.com',
    }
    const edits = setProperty(config, ['github', -1], value, defaultFormattingOptions)
    return { edits, selectText: tokenPlaceholder }
}

const addGitLab: ConfigInsertionFunction = config => {
    const tokenPlaceholder =
        '<personal access token with api scope (https://[your-gitlab-hostname]/profile/personal_access_tokens)>'
    const value: GitLabConnection = {
        url: 'https://gitlab.example.com',
        token: tokenPlaceholder,
    }
    const edits = setProperty(config, ['gitlab', -1], value, defaultFormattingOptions)
    return { edits, selectText: tokenPlaceholder }
}

const addBitbucketServer: ConfigInsertionFunction = config => {
    const tokenPlaceholder =
        '<personal access token with read scope (https://[your-bitbucket-hostname]/plugins/servlet/access-tokens/add)>'
    const value: BitbucketServerConnection = {
        url: 'https://bitbucket.example.com',
        token: tokenPlaceholder,
    }
    const edits = setProperty(config, ['bitbucketServer', -1], value, defaultFormattingOptions)
    return { edits, selectText: tokenPlaceholder }
}

const addAWSCodeCommit: ConfigInsertionFunction = config => {
    const value: AwsCodeCommitConnection = {
        region: '' as any,
        accessKeyID: '',
        secretAccessKey: '',
    }
    const edits = setProperty(config, ['awsCodeCommit', -1], value, defaultFormattingOptions)
    return { edits, selectText: '""', cursorOffset: 1 }
}

const addOtherRepository: ConfigInsertionFunction = config => {
    const urlPlaceholder = '<git clone URL>'
    const value: Repository = {
        url: urlPlaceholder,
        path: '<desired name of repository on Sourcegraph (example: my/repo)>',
    }
    const edits = setProperty(config, ['repos.list', -1], value, defaultFormattingOptions)
    return { edits, selectText: urlPlaceholder }
}

const addGSuiteOIDCAuthProvider: ConfigInsertionFunction = config => {
    const value: OpenIdConnectAuthProvider = {
        type: 'openidconnect',
        issuer: 'https://accounts.google.com',
        clientID: '<see documentation: https://developers.google.com/identity/protocols/OpenIDConnect#getcredentials>',
        clientSecret: '<see same documentation as clientID>',
        requireEmailDomain: "<your company's email domain (example: mycompany.com)>",
    }
    return {
        edits: [
            ...removeProperty(config, ['auth.provider'], defaultFormattingOptions),
            ...setProperty(config, ['auth.providers'], [value], defaultFormattingOptions),
        ],
    }
}

const addSAMLAuthProvider: ConfigInsertionFunction = config => {
    const value: SamlAuthProvider = {
        type: 'saml',
        identityProviderMetadataURL: '<see https://about.sourcegraph.com/docs/server/config/authentication#saml>',
    }
    return {
        edits: [
            ...removeProperty(config, ['auth.provider'], defaultFormattingOptions),
            ...setProperty(config, ['auth.providers'], [value], defaultFormattingOptions),
        ],
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
    { id: 'sourcegraph.site.githubDotCom', label: 'Add GitHub.com repositories', run: addGitHubDotCom },
    {
        id: 'sourcegraph.site.githubEnterprise',
        label: 'Add GitHub Enterprise repositories',
        run: addGitHubEnterprise,
    },
    { id: 'sourcegraph.site.addGitLab', label: 'Add GitLab projects', run: addGitLab },
    { id: 'sourcegraph.site.addBitbucketServer', label: 'Add Bitbucket Server repositories', run: addBitbucketServer },
    { id: 'sourcegraph.site.addAWSCodeCommit', label: 'Add AWS CodeCommit repositories', run: addAWSCodeCommit },
    { id: 'sourcegraph.site.otherRepository', label: 'Add other repository', run: addOtherRepository },
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
