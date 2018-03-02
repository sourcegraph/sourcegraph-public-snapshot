import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { Edit, FormattingOptions } from '@sqs/jsonc-parser/lib/format'
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

/**
 * A helper function that modifies site configuration to configure specific
 * common things, such as syncing GitHub repositories.
 */
export type ConfigInsertionFunction = (
    configJSON: string
) => {
    /** The edits to make to the input configuration to insert the new configuration. */
    edits: Edit[]

    /** Select text in inserted JSON. */
    selectText?: string

    /**
     * If set, the selection is an empty selection that begins at the left-hand match of selectText plus this
     * offset. For example, if selectText is "foo" and cursorOffset is 2, then the final selection will be a cursor
     * "|" positioned as "fo|o".
     */
    cursorOffset?: number
}

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

const addSSOViaGSuite: ConfigInsertionFunction = config => {
    const value: OpenIdConnectAuthProvider = {
        issuer: 'https://accounts.google.com',
        clientID: '<see documentation: https://developers.google.com/identity/protocols/OpenIDConnect#getcredentials>',
        clientSecret: '<see same documentation as clientID>',
        requireEmailDomain: "<your company's email domain (example: mycompany.com)>",
    }
    return {
        edits: [
            ...setProperty(config, ['auth.provider'], 'openidconnect', defaultFormattingOptions),
            ...setProperty(config, ['auth.openIDConnect'], value, defaultFormattingOptions),
            ...setProperty(config, ['auth.allowSignup'], false, defaultFormattingOptions),
        ],
        selectText: '"auth.openIDConnect": {',
    }
}

const addSSOViaSAML: ConfigInsertionFunction = config => {
    const value: SamlAuthProvider = {
        identityProviderMetadataURL: '<see https://about.sourcegraph.com/docs/server/config/authentication#saml>',
        serviceProviderCertificate: '<see https://about.sourcegraph.com/docs/server/config/authentication#saml>',
        serviceProviderPrivateKey: '<see https://about.sourcegraph.com/docs/server/config/authentication#saml>',
    }
    return {
        edits: [
            ...setProperty(config, ['auth.provider'], 'saml', defaultFormattingOptions),
            ...setProperty(config, ['auth.saml'], value, defaultFormattingOptions),
        ],
        selectText: '"auth.saml": {',
    }
}

const addSearchScopeToSettings: ConfigInsertionFunction = config => {
    const value: { name: string; value: string } = {
        name: '<name>',
        value: '<partial query string that will be inserted when the scope is selected>',
    }
    const edits = setProperty(config, ['search.scopes', -1], value, defaultFormattingOptions)
    return { edits, selectText: '<name>' }
}

export interface EditorAction {
    id: string
    label: string
    run: ConfigInsertionFunction
}

export const settingsActions: EditorAction[] = [
    { id: 'sourcegraph.settings.searchScopes', label: 'Add search scope', run: addSearchScopeToSettings },
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
    { id: 'sourcegraph.site.ssoViaGSuite', label: 'Use SSO via Google (G Suite)', run: addSSOViaGSuite },
    { id: 'sourcegraph.site.ssoViaSAML', label: 'Use SSO via SAML', run: addSSOViaSAML },
]

/**
 * Parses out the 'disableTelemetry' key from the JSON site config and returns the inverse.
 * If this config option is not set, or if the configuration has a parse error, default is true.
 */
export function getTelemetryEnabled(text: string): boolean {
    try {
        return !(parseJSON(text) as SiteConfiguration).disableTelemetry
    } catch (err) {
        console.error(err)
        return true
    }
}
