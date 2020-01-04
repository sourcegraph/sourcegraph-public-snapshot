import { Edit, FormattingOptions, JSONPath } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import AmazonIcon from 'mdi-react/AmazonIcon'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import GitIcon from 'mdi-react/GitIcon'
import GitLabIcon from 'mdi-react/GitlabIcon'
import React from 'react'
import awsCodeCommitSchemaJSON from '../../../schema/aws_codecommit.schema.json'
import bitbucketCloudSchemaJSON from '../../../schema/bitbucket_cloud.schema.json'
import bitbucketServerSchemaJSON from '../../../schema/bitbucket_server.schema.json'
import githubSchemaJSON from '../../../schema/github.schema.json'
import gitlabSchemaJSON from '../../../schema/gitlab.schema.json'
import gitoliteSchemaJSON from '../../../schema/gitolite.schema.json'
import otherExternalServiceSchemaJSON from '../../../schema/other_external_service.schema.json'
import phabricatorSchemaJSON from '../../../schema/phabricator.schema.json'
import { PhabricatorIcon } from '../../../shared/src/components/icons'
import * as GQL from '../../../shared/src/graphql/schema'
import { EditorAction } from './configHelpers.js'

/**
 * Metadata associated with adding a given external service.
 * TODO: rename to AddExternalServiceOptions
 */
export interface ExternalServiceKindMetadata {
    kind: GQL.ExternalServiceKind

    /**
     * Title to show in the external service "button"
     */
    title: string

    /**
     * Icon to show in the external service "button"
     */
    icon: React.ComponentType<{ className?: string }>

    /**
     * A short description that will appear in the external service "button" under the title
     */
    shortDescription?: string

    /**
     * Instructions that will appear on the add / edit page
     */
    instructions?: JSX.Element | string

    /**
     * The JSON schema of the external service configuration
     */
    jsonSchema: { $id: string }

    /**
     * Quick configure editor actions
     */
    editorActions?: EditorAction[]

    /**
     * Default display name
     */
    defaultDisplayName: string

    /**
     * Default external service configuration
     */
    defaultConfig: string
}

const defaultFormattingOptions: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

/**
 * editWithComment returns a Monaco edit action that sets the value of a JSON field with a
 * "//" comment annotating the field. The comment is inserted wherever
 * `"COMMENT_SENTINEL": true` appears in the JSON.
 */
function editWithComment(config: string, path: JSONPath, value: any, comment: string): Edit {
    const edit = setProperty(config, path, value, defaultFormattingOptions)[0]
    edit.content = edit.content.replace('"COMMENT_SENTINEL": true', comment)
    return edit
}

const editorActionComments = {
    enablePermissions:
        '// Prerequisite: you must configure GitHub as an OAuth auth provider in the site config (https://docs.sourcegraph.com/admin/auth#github). Otherwise, access to all repositories will be disallowed.',
    enforcePermissionsOAuth: `// Prerequisite: you must first update the site configuration to
    // include GitLab OAuth as an auth provider.
    // See https://docs.sourcegraph.com/admin/auth#gitlab for instructions.`,
    enforcePermissionsSSO: `// Prerequisite: You will need a sudo-level access token. If you can configure
    // GitLab as an OAuth identity provider for Sourcegraph, we recommend that
    // option instead.
    //
    // 1. Ensure the personal access token in this config has admin privileges
    //    (https://docs.gitlab.com/ee/api/#sudo).
    // 2. Update the site configuration to include the SSO auth provider for GitLab
    //    (https://docs.sourcegraph.com/admin/auth).
    // 3. Update the fields below to match the properties of this auth provider
    //    (https://docs.sourcegraph.com/admin/repo/permissions#sudo-access-token).`,
}

const GITHUB_DOTCOM: ExternalServiceKindMetadata = {
    kind: GQL.ExternalServiceKind.GITHUB,
    title: 'GitHub.com',
    icon: GithubCircleIcon,
    jsonSchema: githubSchemaJSON,
    editorActions: [
        {
            id: 'setAccessToken',
            label: 'Set access token',
            run: (config: string) => {
                const value = '<access token>'
                const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                return { edits, selectText: '<access token>' }
            },
        },
        {
            id: 'addOrgRepo',
            label: 'Add repositories in an organization',
            run: (config: string) => {
                const value = '<organization name>'
                const edits = setProperty(config, ['orgs', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<organization name>' }
            },
        },
        {
            id: 'addSearchQueryRepos',
            label: 'Add repositories matching a search query',
            run: (config: string) => {
                const value = '<search query>'
                const edits = setProperty(config, ['repositoryQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<search query>' }
            },
        },
        {
            id: 'addAffiliatedRepos',
            label: 'Add repositories affiliated with token',
            run: (config: string) => {
                const value = 'affiliated'
                const edits = setProperty(config, ['repositoryQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: 'affiliated' }
            },
        },
        {
            id: 'addRepo',
            label: 'Add a single repository',
            run: config => {
                const value = '<owner>/<repository>'
                const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<owner>/<repository>' }
            },
        },
        {
            id: 'excludeRepo',
            label: 'Exclude a repository',
            run: config => {
                const value = { name: '<owner>/<repository>' }
                const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<owner>/<repository>' }
            },
        },
        {
            id: 'enablePermissions',
            label: 'Enforce permissions',
            run: config => {
                const value = {
                    COMMENT_SENTINEL: true,
                }
                const comment = editorActionComments.enablePermissions
                const edit = editWithComment(config, ['authorization'], value, comment)
                return { edits: [edit], selectText: comment }
            },
        },
    ],
    instructions: (
        <div>
            <ol>
                <li>
                    Create a GitHub access token (
                    <a
                        href="https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        instructions
                    </a>
                    ) with <b>repo</b> scope, and set it to be the value of the <code>token</code> field in the
                    configuration below.
                </li>
                <li>
                    Specify which repositories Sourcegraph should index using one of the following fields:
                    <ul>
                        <li>
                            <code>organizations</code>: specify a list of GitHub organizations.
                        </li>
                        <li>
                            <code>repositoryQuery</code>: specify a list of GitHub search queries. Use "affiliated" to
                            specify all repositories associated with the access token.
                        </li>
                        <li>
                            <code>repos</code>: list individual repositories by name.
                        </li>
                    </ul>
                </li>
            </ol>
            <p>
                See{' '}
                <a
                    rel="noopener noreferrer"
                    target="_blank"
                    href="https://docs.sourcegraph.com/admin/external_service/github#configuration"
                >
                    the docs for more advanced options
                </a>{' '}
                or try out one of the buttons below.
            </p>
        </div>
    ),
    defaultDisplayName: 'GitHub',
    defaultConfig: `{
  "url": "https://github.com",
  "token": "<access token>",
  "orgs": []
}`,
}
const GITHUB_ENTERPRISE: ExternalServiceKindMetadata = {
    ...GITHUB_DOTCOM,
    title: 'GitHub Enterprise',
    defaultConfig: `{
  "url": "https://github.example.com",
  "token": "<access token>",
  "orgs": []
}`,
    instructions: (
        <div>
            <ol>
                <li>
                    Set <code>url</code> to be the URL of GitHub Enterprise.
                </li>
                <li>
                    Create a GitHub access token (
                    <a
                        href="https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        instructions
                    </a>
                    ) with <b>repo</b> scope, and set it to be the value of the <code>token</code> field in the
                    configuration below.
                </li>
                <li>
                    Specify which repositories Sourcegraph should index using one of the following fields:
                    <ul>
                        <li>
                            <code>organizations</code>: specify a list of GitHub organizations.
                        </li>
                        <li>
                            <code>repositoryQuery</code>: specify a list of GitHub search queries. Use "affiliated" to
                            specify all repositories associated with the access token.
                        </li>
                        <li>
                            <code>repos</code>: list individual repositories by name.
                        </li>
                    </ul>
                </li>
            </ol>
            <p>
                See{' '}
                <a
                    rel="noopener noreferrer"
                    target="_blank"
                    href="https://docs.sourcegraph.com/admin/external_service/github#configuration"
                >
                    the docs for more advanced options
                </a>{' '}
                or try out one of the buttons below.
            </p>
        </div>
    ),
}
const AWS_EXTERNAL_SERVICE: ExternalServiceKindMetadata = {
    kind: GQL.ExternalServiceKind.AWSCODECOMMIT,
    title: 'AWS CodeCommit repositories',
    icon: AmazonIcon,
    shortDescription: 'Add AWS CodeCommit repositories.',
    jsonSchema: awsCodeCommitSchemaJSON,
    defaultDisplayName: 'AWS CodeCommit',
    defaultConfig: `// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// AWS CodeCommit external service docs: https://docs.sourcegraph.com/admin/external_service/aws_codecommit#configuration
{
"accessKeyID": "<access key id>",
"secretAccessKey": "<secret access key>",
"region": "<region>",

// Git credentials for cloning an AWS CodeCommit repository over https
// See IAM Code Commit auth docs: https://docs.aws.amazon.com/codecommit/latest/userguide/setting-up-gc.html
"gitCredentials": {
"username": "<username>",
"password": "<password>"
},

// Repositories to exclude by name ({"name": "git-codecommit.us-west-1.amazonaws.com/repo-name"})
// or by ARN ({"id": "arn:aws:codecommit:us-west-1:999999999999:name"})
// "exclude": [
//   {
//     "name": "mono-repo"
//   }
// ]
}`,
    editorActions: [
        {
            id: 'setAccessKeyID',
            label: 'Set access key ID',
            run: config => {
                const value = '<access key id>'
                const edits = setProperty(config, ['accessKeyID'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setSecretAccessKey',
            label: 'Set secret access key',
            run: config => {
                const value = '<secret access key>'
                const edits = setProperty(config, ['secretAccessKey'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setRegion',
            label: 'Set region',
            run: config => {
                const value = '<region>'
                const edits = setProperty(config, ['region'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setGitCredentials',
            label: 'Set Git credentials',
            run: config => {
                const value = {
                    username: '<username>',
                    password: '<password>',
                }
                const edits = setProperty(config, ['gitCredentials'], value, defaultFormattingOptions)
                return { edits, selectText: '<username>' }
            },
        },
        {
            id: 'excludeRepo',
            label: 'Exclude a repository',
            run: config => {
                const value = { name: '<owner>/<repository>' }
                const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<owner>/<repository>' }
            },
        },
    ],
}
const BITBUCKET_CLOUD_SERVICE: ExternalServiceKindMetadata = {
    kind: GQL.ExternalServiceKind.BITBUCKETCLOUD,
    title: 'Bitbucket.org repositories',
    icon: BitbucketIcon,
    shortDescription: 'Add Bitbucket Cloud repositories.',
    jsonSchema: bitbucketCloudSchemaJSON,
    defaultDisplayName: 'Bitbucket Cloud',
    defaultConfig: `// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// Bitbucket Cloud external service docs: https://docs.sourcegraph.com/admin/external_service/bitbucket_cloud#configuration
{
"url": "https://bitbucket.org",

// The username the app password belongs to
"username": "<username>",

// An app password (https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html) with read scope over the repositories and teams to be added to Sourcegraph
"appPassword": "<app password>",

// teams: List of teams whose repositories should be selected
// "teams": [
//   "<team name>"
// ],
}`,
    editorActions: [
        {
            id: 'setAppPassword',
            label: 'Set app password',
            run: config => {
                const value = '<app password>'
                const edits = setProperty(config, ['appPassword'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
    ],
}
const BITBUCKET_SERVER_SERVICE: ExternalServiceKindMetadata = {
    kind: GQL.ExternalServiceKind.BITBUCKETSERVER,
    title: 'Bitbucket Server repositories',
    icon: BitbucketIcon,
    shortDescription: 'Add Bitbucket Server repositories.',
    jsonSchema: bitbucketServerSchemaJSON,
    defaultDisplayName: 'Bitbucket Server',
    defaultConfig: `// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// Bitbucket Server external service docs: https://docs.sourcegraph.com/admin/external_service/bitbucket_server#configuration
{
"url": "https://bitbucket.example.com",

// Create a personal access token with read scope at
// https://<bitbucket-hostname>/plugins/servlet/access-tokens/add
"token": "<access token>",

// The username the token belongs to
"username": "<username>",

// SELECTING REPOSITORIES
//
// There are 3 fields used to select repositories for searching and code intel:
//  - repositoryQuery (required)
//  - repos
//  - exclude

// repositoryQuery: List of strings: a special keyword "none" (which disables querying),
// "all" (which selects all repositories visible to the given token), or any repository
// search query parameters (e.g "?name=<repo name>&projectname=<project>&visibility=private")
// See the list of parameters at: https://docs.atlassian.com/bitbucket-server/rest/6.1.2/bitbucket-rest.html#idp355
"repositoryQuery": [],

// repos: Explicit list of repositories to select
// "repos": [
//   "<project/<repository>"
// ],

// exclude: Repositories to exclude (overrides repositories from repositoryQuery and repos)
// "exclude": [
//   {
//     "name": "<project/<repository>"
//   }
// ]
}`,
    editorActions: [
        {
            id: 'setURL',
            label: 'Set Bitbucket Server URL',
            run: config => {
                const value = 'https://bitbucket.example.com'
                const edits = setProperty(config, ['url'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setPersonalAccessToken',
            label: 'Set access token',
            run: config => {
                const value = '<access token>'
                const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addProjectRepos',
            label: 'Add project repositories',
            run: config => {
                const value = '?projectname=<project>'
                const edits = setProperty(config, ['repositoryQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<project>' }
            },
        },
        {
            id: 'addRepo',
            label: 'Add a repository',
            run: config => {
                const value = '<project/<repository>'
                const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<project/<repository>' }
            },
        },
        {
            id: 'excludeRepo',
            label: 'Exclude a repository',
            run: config => {
                const value = { name: '<project/<repository>' }
                const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                return { edits, selectText: '{"name": "<project/<repository>"}' }
            },
        },
        {
            id: 'setSelfSignedCert',
            label: 'Set internal or self-signed certificate',
            run: config => {
                const value = '<certificate>'
                const edits = setProperty(config, ['certificate'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'enablePermissions',
            label: 'Enforce permissions',
            run: config => {
                const value = {
                    COMMENT_SENTINEL: true,
                    identityProvider: { type: 'username' },
                    oauth: {
                        consumerKey: '<consumer key>',
                        signingKey: '<signing key>',
                    },
                    ttl: '3h',
                    hardTTL: '72h',
                }
                const comment =
                    '// Follow setup instructions in https://docs.sourcegraph.com/admin/repo/permissions#bitbucket_server'
                const edit = editWithComment(config, ['authorization'], value, comment)
                return { edits: [edit], selectText: comment }
            },
        },
    ],
}
const GITLAB_SERVICE: ExternalServiceKindMetadata = {
    kind: GQL.ExternalServiceKind.GITLAB,
    title: 'GitLab projects',
    icon: GitLabIcon,
    shortDescription: 'Add GitLab projects.',
    jsonSchema: gitlabSchemaJSON,
    defaultDisplayName: 'GitLab',
    defaultConfig: `// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// GitLab external service docs: https://docs.sourcegraph.com/admin/external_service/gitlab#configuration
{
"url": "https://example.gitlab.com",

// Create a personal access token with api scope at https://[your-gitlab-hostname]/profile/personal_access_tokens
"token": "<access token>",

// SELECTING REPOSITORIES
//
// There are 3 fields used to select repositories for searching and code intel:
//  - projectQuery (required)
//  - projects
//  - exclude

// List of strings, either a special keyword "none" (which disables querying), search query parameters
// such as "?search=sourcegraph", and "?visibility=private".
//
// For getting started, use the "Quick configure" buttons above the editor to build an initial set of
// of queries.
"projectQuery": [
//   "?archived=no\u0026visibility=private" // set this to "none" to disable querying
],

// projects: Project repositories to select.  Supports name: {"name": "group/name"}, or ID: {"id": 42})
// "projects": [
//   { "name": "<group>/<name>" },
//   { "id": <id> }
// ],

// exclude: Project repositories to exclude.  Supports name: {"name": "group/name"}, or ID: {"id": 42})
// "exclude": [
//   { "name": "<group>/<name>" },
//   { "id": <id> }
// ]
}`,
    editorActions: [
        {
            id: 'setURL',
            label: 'Set GitLab URL',
            run: config => {
                const value = 'https://gitlab.example.com'
                const edits = setProperty(config, ['url'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setPersonalAccessToken',
            label: 'Set access token',
            run: config => {
                const value = '<access token>'
                const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setSelfSignedCert',
            label: 'Set internal or self-signed certificate',
            run: config => {
                const value = '<certificate>'
                const edits = setProperty(config, ['certificate'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'syncInternalProjects',
            label: 'Sync internal projects',
            run: config => {
                const value = '?visibility=internal'
                const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'syncPrivateProjects',
            label: 'Sync private projects',
            run: config => {
                const value = '?visibility=private'
                const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'syncPublicProjects',
            label: 'Sync public projects',
            run: config => {
                const value = '?visibility=public'
                const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'syncGroupProjects',
            label: 'Sync group projects',
            run: config => {
                const value = 'groups/<group ID>/projects'
                const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<group ID>' }
            },
        },
        {
            id: 'syncMembershipProjects',
            label: 'Sync all projects the access token user is a member of',
            run: config => {
                const value = '?membership=true'
                const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'syncProjectsMatchingSearch',
            label: 'Sync projects matching search',
            run: config => ({
                edits: setProperty(config, ['projectQuery', -1], '?search=<search query>', defaultFormattingOptions),
                selectText: '<search query>',
            }),
        },
        {
            id: 'enforcePermissionsOAuth',
            label: 'Enforce permissions (OAuth)',
            run: config => {
                const value = {
                    identityProvider: {
                        COMMENT_SENTINEL: true,
                        type: 'oauth',
                    },
                }
                const comment = editorActionComments.enforcePermissionsOAuth
                const edit = editWithComment(config, ['authorization'], value, comment)
                return { edits: [edit], selectText: comment }
            },
        },
        {
            id: 'enforcePermissionsSSO',
            label: 'Enforce permissions (SSO)',
            run: config => {
                const value = {
                    COMMENT_SENTINEL: true,
                    identityProvider: {
                        type: 'external',
                        authProviderID: '<configID field of the auth provider>',
                        authProviderType: '<type field of the auth provider>',
                        gitlabProvider:
                            '<name that identifies the auth provider to GitLab (hover over "gitlabProvider" for docs)>',
                    },
                }
                const comment = editorActionComments.enforcePermissionsSSO
                const edit = editWithComment(config, ['authorization'], value, comment)
                return { edits: [edit], selectText: comment }
            },
        },
        {
            id: 'addProject',
            label: 'Add a project',
            run: config => {
                const value = { name: '<group>/<project>' }
                const edits = setProperty(config, ['projects', -1], value, defaultFormattingOptions)
                return { edits, selectText: '{"name": "<group>/<project>"}' }
            },
        },
        {
            id: 'excludeProject',
            label: 'Exclude a project',
            run: config => {
                const value = { name: '<group>/<project>' }
                const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                return { edits, selectText: '{"name": "<group>/<project>"}' }
            },
        },
    ],
}
const GITOLITE_SERVICE: ExternalServiceKindMetadata = {
    kind: GQL.ExternalServiceKind.GITOLITE,
    title: 'Gitolite repositories',
    icon: GitIcon,
    shortDescription: 'Add Gitolite repositories.',
    jsonSchema: gitoliteSchemaJSON,
    defaultDisplayName: 'Gitolite',
    defaultConfig: `{
// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// Configuration options are documented here:
// https://docs.sourcegraph.com/admin/external_service/gitolite#configuration

"prefix": "gitolite.example.com/",
"host": "git@gitolite.example.com"
}`,
    editorActions: [
        {
            id: 'setHost',
            label: 'Set host',
            run: config => {
                const value = 'git@gitolite.example.com'
                const edits = setProperty(config, ['host'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setPrefix',
            label: 'Set prefix',
            run: config => {
                const value = 'gitolite.example.com/'
                const edits = setProperty(config, ['prefix'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
    ],
}
const PHABRICATOR_SERVICE: ExternalServiceKindMetadata = {
    kind: GQL.ExternalServiceKind.PHABRICATOR,
    title: 'Phabricator connection',
    icon: PhabricatorIcon,
    shortDescription:
        'Associate Phabricator repositories with existing repositories on Sourcegraph. Mirroring is not supported.',
    jsonSchema: phabricatorSchemaJSON,
    defaultDisplayName: 'Phabricator',
    defaultConfig: `{
// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// Configuration options are documented here:
// https://docs.sourcegraph.com/admin/external_service/phabricator#configuration

"url": "https://phabricator.example.com",
"token": "",
"repos": []
}`,
    editorActions: [
        {
            id: 'setPhabricatorURL',
            label: 'Set Phabricator URL',
            run: config => {
                const value = 'https://phabricator.example.com'
                const edits = setProperty(config, ['url'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setAccessToken',
            label: 'Set Phabricator access token',
            run: config => {
                const value = '<Phabricator access token>'
                const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addRepository',
            label: 'Add a repository',
            run: config => {
                const value = {
                    callsign: '<Phabricator repository callsign>',
                    path: '<Sourcegraph repository full name>',
                }
                const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<Phabricator repository callsign>' }
            },
        },
    ],
}
const OTHER_SERVICE: ExternalServiceKindMetadata = {
    kind: GQL.ExternalServiceKind.OTHER,
    title: 'Single Git repositories',
    icon: GitIcon,
    shortDescription: 'Add single Git repositories by clone URL.',
    jsonSchema: otherExternalServiceSchemaJSON,
    defaultDisplayName: 'Git repositories',
    defaultConfig: `{
// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// Configuration options are documented here:
// https://docs.sourcegraph.com/admin/external_service/other#configuration

// Supported URL schemes are: http, https, git and ssh
"url": "https://git.example.com",

// Repository clone paths may be relative to the url (preferred) or absolute.
"repos": []
}`,
    editorActions: [
        {
            id: 'setURL',
            label: 'Set Git host URL',
            run: config => {
                const value = 'https://git.example.com'
                const edits = setProperty(config, ['url'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addRepo',
            label: 'Add a repository',
            run: config => {
                const value = 'path/to/repository'
                const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
    ],
}

const EZ_GITLAB_DOTCOM: ExternalServiceKindMetadata = {
    ...GITLAB_SERVICE,
    shortDescription: undefined,
    instructions: (
        <div>
            <ol>
                <li>
                    Create a GitLab access token (
                    <a
                        href="https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#creating-a-personal-access-token"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        instructions
                    </a>
                    ) with{' '}
                    <b>
                        <code>repo</code>
                    </b>{' '}
                    scope, and set it to be the value of the <code>token</code> field in the configuration below.
                </li>
                <li>
                    Specify which projects on GitLab should be cloned with the <code>projectQuery</code> field, which
                    contains a list of strings specifying REST API calls to GitLab that return a list of projects.
                    <ul>
                        <li>
                            Use <code>projects?membership=true&archived=no</code> to select all unarchived projects of
                            which the token's user is a member.
                        </li>
                        <li>
                            Use <code>groups/&lt;mygroup&gt;/projects</code> to select all projects in a group.
                        </li>
                        <li>
                            Alternatively, list individual projects by name or ID with the <code>projects</code> field.
                        </li>
                    </ul>
                </li>
            </ol>
        </div>
    ),
    editorActions: [
        {
            id: 'setAccessToken',
            label: 'Set access token',
            run: (config: string) => {
                const value = '<access token>'
                const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addGroupProjects',
            label: 'Add projects in a group',
            run: (config: string) => {
                const value = 'groups/<my group>/projects'
                const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<my group>' }
            },
        },
        {
            id: 'addMemberProjects',
            label: "Add projects that have the token's user as member",
            run: (config: string) => {
                const value = 'projects?membership=true&archived=no'
                const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addIndividualProjectByName',
            label: 'Add single project by name',
            run: (config: string) => {
                const value = { name: '<group>/<name>' }
                const edits = setProperty(config, ['projects', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<group>/<name>' }
            },
        },
        {
            id: 'addIndividualProjectByID',
            label: 'Add single project by ID',
            run: (config: string) => {
                const value = { id: 123 }
                const edits = setProperty(config, ['projects', -1], value, defaultFormattingOptions)
                return { edits, selectText: '123' }
            },
        },
    ],
    title: 'GitLab.com',
    defaultConfig: `{
  "url": "https://gitlab.com",
  "token": "<access token>",
  "projectQuery": [
    "projects?membership=true&archived=no"
  ]
}`,
}

const EZ_GITLAB_SELFMANAGED = {
    ...EZ_GITLAB_DOTCOM,
    title: 'GitLab Self-Managed',
}

const EZ_BITBUCKET_DOTORG: ExternalServiceKindMetadata = {
    ...BITBUCKET_CLOUD_SERVICE,
    shortDescription: undefined,
    instructions: (
        <div>
            <ol>
                <li>
                    Create a Bitbucket app password (
                    <a
                        href="https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        instructions
                    </a>
                    ) with{' '}
                    <b>
                        <code>read</code>
                    </b>{' '}
                    scope over your repositories and teams. Set it to be the value of the <code>appPassword</code> field
                    in the configuration below.
                </li>
                <li>
                    Set the <code>username</code> field to be the username corresponding to <code>appPassword</code>.
                </li>
                <li>
                    Set the <code>teams</code> field to be the list of teams whose repositories Sourcegraph should
                    index.
                </li>
            </ol>
        </div>
    ),
    editorActions: [
        {
            id: 'setAppPassword',
            label: 'Set app password',
            run: config => {
                const value = '<app password>'
                const edits = setProperty(config, ['appPassword'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setUsername',
            label: 'Set username',
            run: config => {
                const value = '<username to which the app password belongs>'
                const edits = setProperty(config, ['username'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addTeamRepositories',
            label: 'Add repositories belonging to team',
            run: config => {
                const value = '<team>'
                const edits = setProperty(config, ['teams', -1], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
    ],
    title: 'Bitbucket.org',
    defaultConfig: `{
  "url": "https://bitbucket.org",
  "appPassword": "<app password>",
  "username": "<username to which the app password belongs>",
  "teams": [
  ]
}`,
}

const EZ_BITBUCKET_SERVER: ExternalServiceKindMetadata = {
    ...BITBUCKET_SERVER_SERVICE,
    title: 'Bitbucket Server',
    shortDescription: undefined,
    instructions: (
        <div>
            <ol>
                <li>
                    In the configuration below, set <code>url</code> to the URL of Bitbucket Server.
                </li>
                <li>
                    Create a personal access token (
                    <a
                        href="https://confluence.atlassian.com/bitbucketserver/personal-access-tokens-939515499.html"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        instructions
                    </a>
                    ) with <code>read</code> scope.
                </li>
                <li>
                    Set <code>username</code> to the user that created the personal access token.
                </li>
                <li>
                    Specify which repositories Sourcegraph should clone using the following fields.
                    <ul>
                        <li>
                            <code>repositoryQuery</code>: a list of strings that are one of the following:
                            <ul>
                                <li>
                                    <code>"all"</code> selects all repositories visible to the token
                                </li>
                                <li>
                                    A query string like{' '}
                                    <code>"{'?name=<repo name>&projectname=<project>&visibility=private'}"</code> that
                                    specifies search query parameters. See{' '}
                                    <a
                                        href="https://docs.atlassian.com/bitbucket-server/rest/6.1.2/bitbucket-rest.html#idp355"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                    >
                                        the full list of parameters
                                    </a>
                                    .
                                </li>
                                <li>
                                    <code>"none"</code> selects no repositories (should only be used if you are listing
                                    repositories one-by-one)
                                </li>
                            </ul>
                        </li>
                        <li>
                            <code>repos</code>: a list of single repositories
                        </li>
                        <li>
                            <code>exclude</code>: a list of repositories or repository name patterns to exclude
                        </li>
                        <li>
                            <code>excludePersonalRepositories</code>: if true, excludes personal repositories from being
                            indexed
                        </li>
                    </ul>
                </li>
            </ol>
            <p>
                See{' '}
                <a
                    rel="noopener noreferrer"
                    target="_blank"
                    href="https://docs.sourcegraph.com/admin/external_service/bitbucket_server#configuration"
                >
                    the docs for more advanced options
                </a>
            </p>
        </div>
    ),
    defaultConfig: `{
  "url": "https://bitbucket.example.com",
  "token": "<access token>",
  "username": "<username that created access token>",
  "repositoryQuery": [
    "all"
  ]
}`,
    editorActions: [
        {
            id: 'setURL',
            label: 'Set Bitbucket Server URL',
            run: config => {
                const value = 'https://bitbucket.example.com'
                const edits = setProperty(config, ['url'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setPersonalAccessToken',
            label: 'Set access token',
            run: config => {
                const value = '<access token>'
                const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setUsername',
            label: 'Set username',
            run: config => {
                const value = '<username that created access token>'
                const edits = setProperty(config, ['username'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addProjectRepos',
            label: 'Add repositories in a project',
            run: config => {
                const value = '?projectname=<project>'
                const edits = setProperty(config, ['repositoryQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<project>' }
            },
        },
        {
            id: 'addRepo',
            label: 'Add individual repository',
            run: config => {
                const value = '<project/<repository>'
                const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<project/<repository>' }
            },
        },
        {
            id: 'setSelfSignedCert',
            label: 'Set internal or self-signed certificate',
            run: config => {
                const value = '<certificate>'
                const edits = setProperty(config, ['certificate'], value, defaultFormattingOptions)
                return { edits, selectText: value }
            },
        },
    ],
}

const EZ_AWS_CODECOMMIT = {
    ...AWS_EXTERNAL_SERVICE,
    shortDescription: undefined,
    instructions: (
        <div>
            <ol>
                <li>
                    Obtain your AWS Secret Access Key and key ID:
                    <ul>
                        <li>Log in to your AWS Management Console.</li>
                        <li>Click your username in the upper right corner of the page.</li>
                        <li>Click on "My Security Credentials".</li>
                        <li>Scroll to the section, "Access keys for CLI, SDK, & API access".</li>
                        <li>
                            Use an existing key ID if you still have access to its Secret Access Key. Otherwise, click
                            "Create access key" to create a new key. Record the Secret Access Key and key ID in a safe
                            place.
                        </li>
                        <li>
                            Set <code>accessKeyID</code> and <code>secretAccessKey</code> in the configuration below to
                            the access key ID and Secret Access Key.
                        </li>
                    </ul>
                </li>
                <li>
                    Set the region to your AWS region. The region (e.g., <code>us-west-2</code>) should be visible in
                    the URL when you visit AWS CodeCommit. You can visit AWS CodeCommit by logging into AWS, clicking on
                    "Services" in the top navbar, and clicking on "CodeCommit".
                </li>
                <li>
                    Create Git credentials for AWS CodeCommit (
                    <a
                        href="https://docs.aws.amazon.com/codecommit/latest/userguide/setting-up-gc.html#setting-up-gc-iam"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        instructions
                    </a>
                    ) and set these in the <code>gitCredentials</code> field.
                </li>
                <li>
                    You can optionally exclude repositories using the <code>exclude</code> field.
                </li>
            </ol>
        </div>
    ),
    defaultConfig: `{
  "accessKeyID": "<access key id>",
  "secretAccessKey": "<secret access key>",
  "region": "<region>",
  "gitCredentials": {
    "username": "<username>",
    "password": "<password>"
  }
}`,
}

const EZ_GITOLITE = {
    ...GITOLITE_SERVICE,
    title: 'Gitolite',
    shortDescription: undefined,
    instructions: (
        <div>
            <ol>
                <li>
                    In the configuration below, set <code>host</code> to be the username and host of the Gitolite
                    server.
                </li>
                <li>
                    Set the <code>prefix</code> field to the prefix you desire for the repository names on Sourcegraph.
                    This is typically the hostname of the Gitolite server.
                </li>
            </ol>
            <p>
                See{' '}
                <a
                    rel="noopener noreferrer"
                    target="_blank"
                    href="https://docs.sourcegraph.com/admin/external_service/gitolite#configuration"
                >
                    the docs for more advanced options
                </a>
                .
            </p>
        </div>
    ),
    defaultConfig: `{
  "host": "git@gitolite.example.com",
  "prefix": "gitolite.example.com/"
}`,
}

const EZ_GIT = {
    ...OTHER_SERVICE,
    title: 'Generic Git host',
    shortDescription: undefined,
    instructions: (
        <div>
            <ol>
                <li>
                    In the configuration below, set <code>url</code> to be the URL of your Git host.
                </li>
                <li>
                    Add the paths of the repositories you wish to index to the <code>repos</code> field. These will be
                    appended to the host URL to obtain the repository clone URLs.
                </li>
            </ol>
        </div>
    ),
    defaultConfig: `{
  "url": "https://git.example.com",
  "repos": []
}`,
}

export const onboardingExternalServices: Record<string, ExternalServiceKindMetadata> = {
    github: GITHUB_DOTCOM,
    ghe: GITHUB_ENTERPRISE,
    gitlabcom: EZ_GITLAB_DOTCOM,
    gitlab: EZ_GITLAB_SELFMANAGED,
    bitbucket: EZ_BITBUCKET_DOTORG,
    bitbucketserver: EZ_BITBUCKET_SERVER,
    aws_codecommit: EZ_AWS_CODECOMMIT,
    gitolite: EZ_GITOLITE,
    git: EZ_GIT,
}

export const externalServices: Record<string, ExternalServiceKindMetadata> = {
    github: GITHUB_DOTCOM,
    ghe: GITHUB_ENTERPRISE,
    bitbucket: BITBUCKET_CLOUD_SERVICE,
    bitbucketserver: BITBUCKET_SERVER_SERVICE,
    gitlab: GITLAB_SERVICE,
    gitolite: GITOLITE_SERVICE,
    phabricator: PHABRICATOR_SERVICE,
    git: OTHER_SERVICE,
    aws: AWS_EXTERNAL_SERVICE,
}

export const nonCodeHostExternalServices: Record<string, ExternalServiceKindMetadata> = {
    phabricator: PHABRICATOR_SERVICE,
}

export const defaultExternalServices: Record<GQL.ExternalServiceKind, ExternalServiceKindMetadata> = {
    [GQL.ExternalServiceKind.GITHUB]: GITHUB_DOTCOM,
    [GQL.ExternalServiceKind.BITBUCKETCLOUD]: BITBUCKET_CLOUD_SERVICE,
    [GQL.ExternalServiceKind.BITBUCKETSERVER]: BITBUCKET_SERVER_SERVICE,
    [GQL.ExternalServiceKind.GITLAB]: GITLAB_SERVICE,
    [GQL.ExternalServiceKind.GITOLITE]: GITOLITE_SERVICE,
    [GQL.ExternalServiceKind.PHABRICATOR]: PHABRICATOR_SERVICE,
    [GQL.ExternalServiceKind.OTHER]: OTHER_SERVICE,
    [GQL.ExternalServiceKind.AWSCODECOMMIT]: AWS_EXTERNAL_SERVICE,
}
